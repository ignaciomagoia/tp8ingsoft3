#!/bin/sh
set -eu

: "${VITE_API_URL:=http://localhost:8080}"
: "${REACT_APP_API_URL:=${VITE_API_URL}}"

cat <<EOC > /usr/share/nginx/html/env-config.js
window.__ENV__ = window.__ENV__ || {};
window.__ENV__.VITE_API_URL = "${VITE_API_URL}";
window.__ENV__.REACT_APP_API_URL = "${REACT_APP_API_URL}";
EOC

exec nginx -g "daemon off;"
