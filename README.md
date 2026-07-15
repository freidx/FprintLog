# FprintLog

## Starting as service

1. Build the application.

`go build -o FprintLog`

2. Move compiled binary to local user folder and ensure it's executable.

`mkdir -p ~/.local/bin && cp FprintLog ~/.local/bin && chmod +x ~/.local/bin/FprintLog`

3. Create user service to run FprintLog.

`mkdir -p ~/.config/systemd/user`

```bash
cat << 'EOF' > ~/.config/systemd/user/FprintLog.service
[Unit]
Description=keeping a log of failed and successful fingerprint events
After=graphical-session.target

[Service]
ExecStart=%h/.local/bin/FprintLog
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
EOF
```

4. Start the service.

`systemctl --user daemon-reload && systemctl --user enable --now FprintLog.service`

5. Check it's running.

Status: `systemctl --user status FprintLog.service`

Live Log: `journalctl --user -u FprintLog.service -f`

## Notes
this will also log failed authentications outside of login (e.g fprint being used for sudo)