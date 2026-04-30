# totp

A minimal TOTP authenticator for Linux. A background daemon keeps your secrets encrypted on disk and generates codes on demand. A CLI lets you import accounts and copy codes to the clipboard. An optional Quickshell panel gives you a searchable visual interface with live countdown timers.

![Quickshell demo](assets/demo.gif)

---

## Getting started

### 1. Build

```bash
go build -o totp .
```

### 2. Start the daemon

```bash
./totp daemon
```

On the very first run you will be asked once:

```
Set a passphrase for your master key (press Enter to generate one randomly):
```

- **Press Enter** — a random key is generated for you. Nothing to remember.
- **Type a passphrase** — the key is derived from your passphrase and stored. The passphrase itself is never saved and will never be asked again.

After this one-time setup the daemon starts silently on every subsequent run.

### 3. Import your accounts

See [Importing accounts](#importing-accounts) below.

### 4. Get a code

```bash
# list your accounts to find the ID
./totp list

# copy the code for an account to the clipboard
./totp copy <account-id>
```

---

## Running the daemon with systemd

To have the daemon start automatically with your session, create `~/.config/systemd/user/totp.service`:

```ini
[Unit]
Description=TOTP authenticator daemon
After=default.target

[Service]
ExecStart=/path/to/totp daemon
Restart=on-failure
RestartSec=3

[Install]
WantedBy=default.target
```

Then enable and start it:

```bash
systemctl --user daemon-reload
systemctl --user enable --now totp.service
```

---

## Data storage

Your accounts and master key are stored in your home directory:

| File                               | Contents                                |
| ---------------------------------- | --------------------------------------- |
| `~/.local/share/totp/accounts.enc` | Encrypted account store                 |
| `~/.local/share/totp/master.key`   | Master key (readable only by your user) |

Both paths respect `XDG_DATA_HOME` if set. Your TOTP secrets are encrypted at rest and never leave the daemon — the CLI only ever receives account names and metadata, never secrets.

---

## Importing accounts

### From a QR code image

```bash
./totp import-image ~/Downloads/qr.png
```

Both Google Authenticator export QR codes and standard single-account QR codes are supported. This requires `zbarimg` to be installed:

**Arch Linux**

```bash
sudo pacman -S zbar
```

**Debian / Ubuntu**

```bash
sudo apt install zbar-tools
```

**Fedora**

```bash
sudo dnf install zbar
```

**macOS (Homebrew)**

```bash
brew install zbar
```

### From a URI

If you have an `otpauth://totp/` URI — for example from a website's manual entry option — you can import it directly:

```bash
./totp import-text 'otpauth://totp/Example:alice@example.com?secret=JBSWY3DPEHPK3PXP&issuer=Example'
```

### Re-importing

Importing is safe to run multiple times. Existing accounts are never deleted; if the same account is imported again it is updated in place.

---

## CLI reference

```bash
totp daemon                  # start the daemon (foreground)
totp status                  # check the daemon is alive
totp list                    # list accounts as JSON
totp import-image <path>     # import from a QR code image
totp import-text <uri>       # import from an otpauth:// URI
totp copy <id>               # generate a code and copy it to the clipboard
```

---

## Clipboard

`totp copy` automatically uses the right clipboard tool for your session:

| Session          | Tool used |
| ---------------- | --------- |
| Wayland          | `wl-copy` |
| X11              | `xclip`   |
| Neither detected | `wl-copy` |

If you run the daemon as a systemd user service, clipboard access works automatically as long as the service starts after your graphical session.

---

## Quickshell panel

A visual panel is included in `ui/shell.qml`. It requires [Quickshell](https://quickshell.outfoxxed.me).

```bash
qs -p /path/to/totp/ui
```

The panel opens on the right side of the screen and shows all your accounts with a live countdown arc. Type to filter by name, use arrow keys or `j`/`k` to navigate, and press Enter, Space, or click to copy a code. The panel closes automatically after copying.

The Quickshell panel is optional — the daemon and CLI work fully on their own.

---

## Dependencies

| Dependency                                    | Required for                                   |
| --------------------------------------------- | ---------------------------------------------- |
| Go 1.21+                                      | building the project                           |
| `zbarimg` (zbar)                              | importing QR code images                       |
| `zenity`                                      | file picker in the Quickshell panel (optional) |
| `wl-copy` (wl-clipboard)                      | clipboard on Wayland                           |
| `xclip`                                       | clipboard on X11                               |
| [Quickshell](https://quickshell.outfoxxed.me) | the visual panel (optional)                    |
