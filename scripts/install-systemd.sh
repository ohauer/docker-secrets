#!/bin/sh
# Install secrets-sync as a systemd service

set -e

SCRIPT_NAME="install-systemd.sh"
BINARY_SRC="bin/secrets-sync"
BINARY_DEST="/usr/local/bin/secrets-sync"
UNIT_FILE_SRC="examples/systemd/secrets-sync.service"
UNIT_FILE_DEST="/etc/systemd/system/secrets-sync.service"
ENV_FILE_SRC="examples/systemd/secrets-sync.env.example"
ENV_FILE_DEST="/etc/default/secrets-sync"
CONFIG_DIR="/etc/secrets-sync"
MAN_PAGE_SRC="docs/secrets-sync.1"
MAN_PAGE_DEST="/usr/share/man/man1/secrets-sync.1"
DOC_DIR="/usr/share/doc/secrets-sync"

log_message() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') ${SCRIPT_NAME} - $1"
}

check_root() {
    if [ "$(id -u)" -ne 0 ]; then
        log_message "ERROR: This script must be run as root"
        exit 1
    fi
}

check_binary() {
    if [ ! -f "${BINARY_SRC}" ]; then
        log_message "ERROR: Binary not found at ${BINARY_SRC}"
        log_message "Run 'make build' first"
        exit 1
    fi
}

install_binary() {
    log_message "Installing binary to ${BINARY_DEST}"
    cp "${BINARY_SRC}" "${BINARY_DEST}"
    chmod 755 "${BINARY_DEST}"
    log_message "Binary installed successfully"
}

install_unit_file() {
    log_message "Installing systemd unit file to ${UNIT_FILE_DEST}"
    cp "${UNIT_FILE_SRC}" "${UNIT_FILE_DEST}"
    chmod 644 "${UNIT_FILE_DEST}"
    log_message "Unit file installed successfully"
}

create_config_dir() {
    if [ ! -d "${CONFIG_DIR}" ]; then
        log_message "Creating config directory ${CONFIG_DIR}"
        mkdir -p "${CONFIG_DIR}"
        chmod 755 "${CONFIG_DIR}"
    fi
}

create_user() {
    if ! id -u secrets-sync >/dev/null 2>&1; then
        log_message "Creating secrets-sync system user and group"
        useradd -r -s /bin/false -d /nonexistent -c "Secrets Sync Service" secrets-sync
        log_message "User and group created successfully"
    else
        log_message "User secrets-sync already exists, skipping"
    fi
}

install_env_file() {
    if [ ! -f "${ENV_FILE_DEST}" ]; then
        log_message "Installing environment file to ${ENV_FILE_DEST}"
        cp "${ENV_FILE_SRC}" "${ENV_FILE_DEST}"
        chmod 600 "${ENV_FILE_DEST}"
        log_message "Environment file installed (edit ${ENV_FILE_DEST} to configure)"
    else
        log_message "Environment file already exists at ${ENV_FILE_DEST}, skipping"
    fi
}

generate_config() {
    if [ ! -f "${CONFIG_DIR}/config.yaml" ]; then
        log_message "Generating sample config at ${CONFIG_DIR}/config.yaml"
        "${BINARY_DEST}" init > "${CONFIG_DIR}/config.yaml"
        chmod 600 "${CONFIG_DIR}/config.yaml"
        log_message "Sample config generated (edit ${CONFIG_DIR}/config.yaml to configure)"
    else
        log_message "Config file already exists at ${CONFIG_DIR}/config.yaml, skipping"
    fi
}

install_man_page() {
    if [ -f "${MAN_PAGE_SRC}" ]; then
        log_message "Installing man page to ${MAN_PAGE_DEST}"
        mkdir -p "$(dirname "${MAN_PAGE_DEST}")"
        cp "${MAN_PAGE_SRC}" "${MAN_PAGE_DEST}"
        chmod 644 "${MAN_PAGE_DEST}"
        gzip -f "${MAN_PAGE_DEST}"
        log_message "Man page installed (run 'man secrets-sync')"
    else
        log_message "WARNING: Man page not found at ${MAN_PAGE_SRC}, skipping"
    fi
}

install_documentation() {
    log_message "Installing documentation to ${DOC_DIR}"
    mkdir -p "${DOC_DIR}"

    # Copy documentation files
    for doc in docs/*.md; do
        if [ -f "$doc" ]; then
            cp "$doc" "${DOC_DIR}/"
        fi
    done

    # Copy README and LICENSE
    [ -f "README.md" ] && cp "README.md" "${DOC_DIR}/"
    [ -f "LICENSE" ] && cp "LICENSE" "${DOC_DIR}/"

    chmod -R 644 "${DOC_DIR}"/*
    log_message "Documentation installed"
}

reload_systemd() {
    log_message "Reloading systemd daemon"
    systemctl daemon-reload
}

enable_service() {
    log_message "Enabling secrets-sync service"
    systemctl enable secrets-sync
    log_message "Service enabled (will start on boot)"
}

main() {
    log_message "Starting secrets-sync systemd installation"

    check_root
    check_binary
    create_user
    install_binary
    create_config_dir
    install_unit_file
    install_env_file
    generate_config
    install_man_page
    install_documentation
    reload_systemd
    enable_service

    log_message "Installation complete!"
    log_message ""
    log_message "Next steps:"
    log_message "  1. Create output directories for secrets (e.g., mkdir -p /var/secrets)"
    log_message "  2. Set ownership: chown secrets-sync:secrets-sync /var/secrets"
    log_message "  3. Edit configuration: ${CONFIG_DIR}/config.yaml"
    log_message "  4. Edit environment:   ${ENV_FILE_DEST}"
    log_message "  5. Update ReadWritePaths in ${UNIT_FILE_DEST} to match your paths"
    log_message "  6. Reload systemd:     systemctl daemon-reload"
    log_message "  7. Start service:      systemctl start secrets-sync"
    log_message "  8. Check status:       systemctl status secrets-sync"
    log_message "  9. View logs:          journalctl -u secrets-sync -f"
    log_message " 10. Read manual:        man secrets-sync"
    log_message ""
    log_message "To allow other services to read secrets, add their users to the secrets-sync group:"
    log_message "      usermod -a -G secrets-sync <username>"
}

main "$@"
