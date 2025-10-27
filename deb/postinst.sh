#!/bin/bash
set -e

echo "Running post-installation setup"

/usr/bin/catalogue setup

systemctl daemon-reload || true

systemctl enable catalogue-apt-server.service || true
systemctl restart catalogue-apt-server.service || true

systemctl enable catalogue-daemon.service || true
systemctl restart catalogue-daemon.service || true

echo "Setup complete."

exit 0
