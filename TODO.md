# Evil Merge Detector — Roadmap

## Запуск
- [ ] GitHub-репозиторий: создать публичный репо, запушить код, первый релиз через GoReleaser
- [ ] README.md: GIF-демо, примеры использования, бейджи CI
- [ ] GitHub Action: обёртка для CI/CD (`uses: fimskiy/evilmerge-action@v1`)

## Монетизация
- [ ] SaaS backend: cloud API, история сканирований, dashboard, webhooks (Go + PostgreSQL)
- [ ] GitHub/GitLab App: автосканирование PR, комментарий с результатами
- [ ] Лицензионный гейт: `--api-key` для cloud-фич, free tier без ключа

## Продукт
- [ ] SARIF reporter: интеграция с GitHub Code Scanning
- [ ] `--commit` флаг: детальный отчёт по конкретному merge-коммиту
- [ ] `context.Context`: отмена/таймаут для больших репозиториев
- [ ] Лендинг: сайт с ценами и документацией
