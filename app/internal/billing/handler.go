package billing

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/webhook"

	appsession "github.com/fimskiy/evil-merge-detector/app/internal/session"
	"github.com/fimskiy/evil-merge-detector/app/internal/store"
)

type Handler struct {
	client        *stripe.Client
	priceMonthly  string
	priceYearly   string
	webhookSecret string
	sessionSecret []byte
	db            *store.Store
}

func New(secretKey, priceMonthly, priceYearly, webhookSecret string, sessionSecret []byte, db *store.Store) *Handler {
	return &Handler{
		client:        stripe.NewClient(secretKey),
		priceMonthly:  priceMonthly,
		priceYearly:   priceYearly,
		webhookSecret: webhookSecret,
		sessionSecret: sessionSecret,
		db:            db,
	}
}

func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sess, ok := appsession.Get(r, h.sessionSecret)
	if !ok {
		http.Redirect(w, r, "/auth/github", http.StatusFound)
		return
	}

	priceID := h.priceMonthly
	if r.FormValue("period") == "yearly" {
		priceID = h.priceYearly
	}

	params := &stripe.CheckoutSessionCreateParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionCreateLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String("https://evilmerge.dev/dashboard?upgraded=1"),
		CancelURL:  stripe.String("https://evilmerge.dev/dashboard"),
		Metadata: map[string]string{
			"installation_id": strconv.FormatInt(sess.InstallationID, 10),
		},
	}

	checkoutSess, err := h.client.V1CheckoutSessions.Create(r.Context(), params)
	if err != nil {
		log.Printf("stripe checkout: %v", err)
		http.Error(w, "payment service unavailable", http.StatusServiceUnavailable)
		return
	}

	http.Redirect(w, r, checkoutSess.URL, http.StatusSeeOther)
}

func (h *Handler) Portal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sess, ok := appsession.Get(r, h.sessionSecret)
	if !ok {
		http.Redirect(w, r, "/auth/github", http.StatusFound)
		return
	}

	customerID, err := h.db.GetStripeCustomerID(r.Context(), sess.InstallationID)
	if err != nil || customerID == "" {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	params := &stripe.BillingPortalSessionCreateParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String("https://evilmerge.dev/dashboard"),
	}

	portalSess, err := h.client.V1BillingPortalSessions.Create(r.Context(), params)
	if err != nil {
		log.Printf("stripe portal: %v", err)
		http.Error(w, "billing portal unavailable", http.StatusServiceUnavailable)
		return
	}

	http.Redirect(w, r, portalSess.URL, http.StatusSeeOther)
}

func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	const maxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
		return
	}

	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), h.webhookSecret)
	if err != nil {
		log.Printf("stripe webhook signature: %v", err)
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			log.Printf("stripe webhook parse checkout: %v", err)
			break
		}
		if session.Customer == nil {
			break
		}
		installationID, err := strconv.ParseInt(session.Metadata["installation_id"], 10, 64)
		if err != nil {
			log.Printf("stripe webhook: invalid installation_id %q", session.Metadata["installation_id"])
			break
		}
		if err := h.db.UpdateStripeCustomer(ctx, installationID, session.Customer.ID); err != nil {
			log.Printf("stripe webhook: store customer: %v", err)
		}
		if err := h.db.UpdatePlan(ctx, installationID, "pro"); err != nil {
			log.Printf("stripe webhook: update plan: %v", err)
		}
		log.Printf("stripe: installation %d → pro (customer %s)", installationID, session.Customer.ID)

	case "customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			log.Printf("stripe webhook parse subscription: %v", err)
			break
		}
		if sub.Customer == nil {
			break
		}
		inst, err := h.db.GetInstallationByStripeCustomer(ctx, sub.Customer.ID)
		if err != nil {
			log.Printf("stripe webhook: installation not found for customer %s: %v", sub.Customer.ID, err)
			break
		}
		if err := h.db.UpdatePlan(ctx, inst.InstallationID, "free"); err != nil {
			log.Printf("stripe webhook: downgrade plan: %v", err)
		}
		log.Printf("stripe: installation %d → free (subscription cancelled)", inst.InstallationID)
	}

	w.WriteHeader(http.StatusOK)
}
