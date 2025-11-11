#!/bin/sh
set -e

echo "Waiting for Perses to be ready..."
until curl -sf http://perses:8080/ > /dev/null 2>&1; do
  sleep 2
done
echo "Perses is ready!"

echo "Setting up Prometheus datasource..."
curl -X POST http://perses:8080/api/v1/globaldatasources -H "Content-Type: application/json" -d '{
  "kind": "GlobalDatasource",
  "metadata": {"name": "prometheus-local"},
  "spec": {
    "default": true,
    "plugin": {
      "kind": "PrometheusDatasource",
      "spec": {"directUrl": "http://localhost:9090"}
    }
  }
}' 2>/dev/null && echo "✓ Datasource created" || echo "→ Datasource may already exist"

echo "Creating kite-dev project..."
curl -X POST http://perses:8080/api/v1/projects -H "Content-Type: application/json" -d '{
  "kind": "Project",
  "metadata": {"name": "kite-dev"},
  "spec": {"display": {"name": "Kite Development"}}
}' 2>/dev/null && echo "✓ Project created" || echo "→ Project may already exist"

echo "Loading Kite dashboard..."
if curl -sf http://perses:8080/api/v1/projects/kite-dev/dashboards/kite-overview > /dev/null 2>&1; then
  echo "→ Dashboard already exists, skipping"
else
  curl -X POST http://perses:8080/api/v1/projects/kite-dev/dashboards \
    -H "Content-Type: application/json" \
    -d @/dashboard/kite-dashboard.json 2>/dev/null && \
    echo "✓ Dashboard created" || echo "✗ Failed to create dashboard"
fi

echo ""
echo "🎉 Perses setup complete!"
echo "📊 View dashboard at: http://localhost:3000/projects/kite-dev/dashboards/kite-overview"

