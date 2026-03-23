#!/usr/bin/env bash
set -euo pipefail

REPO="fimskiy/evil-merge-detector"
BINARY="evilmerge"
RELEASES_URL="https://github.com/${REPO}/releases/download"
API_URL="https://api.github.com/repos/${REPO}/releases/latest"
readonly WORK_DIR="${RUNNER_TEMP:-/tmp}/evil-merge-$$"

cleanup() { rm -rf "$WORK_DIR"; }

die() {
  echo "::error::$1"
  exit 1
}

# Escape % \n \r for use in GitHub Actions workflow commands.
# % must go first, otherwise the encoded sequences get double-encoded.
sanitize() {
  local s="$1"
  s="${s//'%'/'%25'}"
  s="${s//$'\n'/'%0A'}"
  s="${s//$'\r'/'%0D'}"
  printf '%s' "$s"
}

# sha256sum on Linux, shasum on macOS — same job, different names.
verify_checksum() {
  local dir="$1" target="$2"

  local checkline
  checkline=$(grep -F "$target" "${dir}/checksums.txt" | head -1)
  [[ -n "$checkline" ]] || { echo "::warning::${target} not found in checksums.txt"; return 1; }

  echo "$checkline" > "${dir}/verify.txt"

  if command -v sha256sum &>/dev/null; then
    (cd "$dir" && sha256sum --check verify.txt >/dev/null)
  elif command -v shasum &>/dev/null; then
    (cd "$dir" && shasum -a 256 --check verify.txt >/dev/null)
  else
    echo "::warning::No sha256 tool found, skipping checksum verification"
  fi
}

validate_inputs() {
  local val
  for val in "${INPUT_SEVERITY:-}" "${INPUT_FAIL_ON:-warning}"; do
    case "$val" in
      ""|info|warning|critical) ;;
      *) die "Invalid severity '${val}'. Expected: info, warning, critical" ;;
    esac
  done

  if [[ -n "${INPUT_SINCE:-}" && ! "${INPUT_SINCE}" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]]; then
    die "Invalid --since value '${INPUT_SINCE}'. Expected: YYYY-MM-DD"
  fi
}

detect_platform() {
  local raw_os raw_arch
  raw_os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  raw_arch="$(uname -m)"

  case "$raw_os" in
    linux)                OS="linux"   ;;
    darwin)               OS="darwin"  ;;
    mingw*|msys*|cygwin*) OS="windows" ;;
    *) die "Unsupported OS: ${raw_os}" ;;
  esac

  case "$raw_arch" in
    x86_64|amd64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) die "Unsupported architecture: ${raw_arch}" ;;
  esac
}

resolve_version() {
  local version="$1"

  # Skip validation for the prebuilt test mode
  [[ "$version" == "prebuilt" ]] && { printf '%s' "$version"; return 0; }

  if [[ "$version" == "latest" ]]; then
    command -v jq &>/dev/null || die "jq is required to resolve latest version"
    version=$(curl -sL --fail --max-time 30 "$API_URL" \
      | jq -r '.tag_name // empty') \
      || die "Failed to fetch latest release from GitHub API (rate-limited or network error)"
    [[ -n "$version" ]] || die "GitHub API returned no tag_name for latest release"
  fi

  version="${version#v}"

  [[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+([a-zA-Z0-9._-]*)?$ ]] \
    || die "Unexpected version format: '${version}'"

  printf '%s' "$version"
}

install_binary() {
  local version="$1"

  # prebuilt = CI/dev mode, binary already in PATH
  if [[ "$version" == "prebuilt" ]]; then
    command -v "${BINARY}" &>/dev/null \
      || die "INPUT_VERSION=prebuilt but '${BINARY}' not found in PATH"
    echo "Using prebuilt ${BINARY}: $(command -v "${BINARY}")"
    return 0
  fi

  local ext="tar.gz"
  [[ "$OS" == "windows" ]] && ext="zip"

  local archive="${BINARY}_${version}_${OS}_${ARCH}.${ext}"
  local base_url="${RELEASES_URL}/v${version}"

  echo "Installing ${BINARY} v${version} (${OS}/${ARCH})..."

  curl -sL --fail --max-time 120 -o "${WORK_DIR}/${archive}" "${base_url}/${archive}" \
    || die "Download failed: ${base_url}/${archive}"

  if curl -sL --fail --max-time 30 -o "${WORK_DIR}/checksums.txt" "${base_url}/checksums.txt" 2>/dev/null; then
    verify_checksum "$WORK_DIR" "$archive" \
      || die "Checksum verification failed for ${archive}"
  else
    echo "::warning::Checksum file not available, skipping verification"
  fi

  if [[ "$ext" == "zip" ]]; then
    unzip -q "${WORK_DIR}/${archive}" -d "$WORK_DIR"
  else
    tar -xzf "${WORK_DIR}/${archive}" -C "$WORK_DIR"
  fi

  local bin_name="${BINARY}"
  [[ "$OS" == "windows" ]] && bin_name="${BINARY}.exe"

  local install_dir="${HOME}/.local/bin"
  mkdir -p "$install_dir"
  chmod +x "${WORK_DIR}/${bin_name}"
  mv "${WORK_DIR}/${bin_name}" "${install_dir}/${bin_name}"

  echo "${install_dir}" >> "$GITHUB_PATH"
  export PATH="${install_dir}:${PATH}"

  echo "Installed ${BINARY} v${version}"
}

# Builds the base scan args (severity/branch/since) into the named array.
# Callers append --format and other flags before running.
build_scan_args() {
  local -n _args=$1
  _args=("scan")
  [[ -n "${INPUT_SEVERITY:-}" ]] && _args+=("--severity=${INPUT_SEVERITY}")
  [[ -n "${INPUT_BRANCH:-}"   ]] && _args+=("--branch=${INPUT_BRANCH}")
  [[ -n "${INPUT_SINCE:-}"    ]] && _args+=("--since=${INPUT_SINCE}")
}

run_scan() {
  local args
  build_scan_args args
  args+=("--format=json" "--fail-on=${INPUT_FAIL_ON:-warning}")

  echo "Running: ${BINARY} ${args[*]} ."

  local exit_code=0
  "${BINARY}" "${args[@]}" . > "${WORK_DIR}/results.json" 2>"${WORK_DIR}/scan.log" \
    || exit_code=$?

  [[ -s "${WORK_DIR}/scan.log" ]] && cat "${WORK_DIR}/scan.log" >&2

  # CLI only supports one output format per run
  if [[ "${INPUT_UPLOAD_SARIF:-false}" == "true" ]]; then
    local sarif_args
    build_scan_args sarif_args
    sarif_args+=("--format=sarif")

    if ! "${BINARY}" "${sarif_args[@]}" . > evil-merge-results.sarif 2>"${WORK_DIR}/sarif.log"; then
      cat "${WORK_DIR}/sarif.log" >&2
      echo "::warning::Failed to generate SARIF output"
    fi
  fi

  return "$exit_code"
}

emit_annotations() {
  local results="${WORK_DIR}/results.json"
  [[ -s "$results" ]] || return 0

  if ! command -v jq &>/dev/null; then
    echo "::warning::jq not found, skipping annotations"
    return 0
  fi

  local count
  count=$(jq -r '.reports // [] | length' "$results" 2>/dev/null || echo "0")
  [[ "$count" != "0" && "$count" != "null" ]] || return 0

  # Sanitize every field to block workflow command injection via crafted paths/messages
  local sev file detail commit msg
  while IFS=$'\t' read -r sev file detail commit msg; do
    file="$(sanitize "$file")"
    detail="$(sanitize "$detail")"
    commit="$(sanitize "$commit")"
    msg="$(sanitize "$msg")"

    case "$sev" in
      2) echo "::error file=${file},title=Evil Merge (CRITICAL)::${detail} [${commit}] ${msg}" ;;
      1) echo "::warning file=${file},title=Evil Merge (WARNING)::${detail} [${commit}] ${msg}" ;;
      *) echo "::notice file=${file},title=Evil Merge (INFO)::${detail} [${commit}] ${msg}" ;;
    esac
  done < <(jq -r '
    .reports[]? |
    .commit_hash as $hash |
    .message as $msg |
    .evil_changes[]? |
    [
      (.severity | tostring),
      .file_path,
      .detail,
      ($hash | .[0:8]),
      ($msg | split("\n") | .[0] | .[0:60])
    ] | join("\t")
  ' "$results" 2>/dev/null)
}

emit_summary() {
  local results="${WORK_DIR}/results.json"
  local summary="$GITHUB_STEP_SUMMARY"

  if [[ ! -s "$results" ]]; then
    printf '### Evil Merge Detector\n\nNo evil merges detected.\n' >> "$summary"
    return 0
  fi

  if ! command -v jq &>/dev/null; then
    printf '### Evil Merge Detector\n\n> Results unavailable: `jq` is not installed on this runner.\n' >> "$summary"
    return 0
  fi

  local count total_changes
  count=$(jq -r '.reports // [] | length' "$results" 2>/dev/null || echo "0")

  if [[ "$count" == "0" || "$count" == "null" ]]; then
    printf '### Evil Merge Detector\n\nNo evil merges detected.\n' >> "$summary"
    return 0
  fi

  total_changes=$(jq -r '[.reports[]?.evil_changes // [] | length] | add // 0' "$results" 2>/dev/null || echo "0")

  {
    echo "### Evil Merge Detector"
    echo ""
    echo "Found **${count}** evil merge commit(s) with **${total_changes}** suspicious change(s)."
    echo ""
    echo "| Severity | Commit | File | Detail |"
    echo "|----------|--------|------|--------|"
    jq -r '
      .reports[]? |
      .commit_hash as $hash |
      .evil_changes[]? |
      (if .severity == 2 then "CRITICAL"
       elif .severity == 1 then "WARNING"
       else "INFO" end) as $sev |
      (.file_path | gsub("\\|"; "\\|")) as $path |
      (.detail   | gsub("\\|"; "\\|")) as $detail |
      "| \($sev) | `\($hash | .[0:8])` | `\($path)` | \($detail) |"
    ' "$results" 2>/dev/null || true
  } >> "$summary"
}

set_outputs() {
  local results="${WORK_DIR}/results.json"
  local count="0" found="false"

  if [[ -s "$results" ]] && command -v jq &>/dev/null; then
    count=$(jq -r '.reports // [] | length' "$results" 2>/dev/null || echo "0")
    if [[ "$count" != "0" && "$count" != "null" ]]; then
      found="true"
    else
      count="0"
    fi
  fi

  echo "found=${found}" >> "$GITHUB_OUTPUT"
  echo "count=${count}" >> "$GITHUB_OUTPUT"
}

main() {
  mkdir -p "$WORK_DIR"
  trap cleanup EXIT

  validate_inputs
  detect_platform

  local version
  version="$(resolve_version "$INPUT_VERSION")"
  install_binary "$version"

  local scan_exit=0
  run_scan || scan_exit=$?

  emit_annotations
  emit_summary
  set_outputs

  exit "$scan_exit"
}

main "$@"
