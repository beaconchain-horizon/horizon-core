# Branch: master — Switch Operational Network

این شاخه شبکه عملیاتی است:
- cmd/switch/ (۳۵ فایل، صنعتی، BNPL، ledger)
- cmd/api/ (Backend :8080)
- cmd/angel-node/ (۷ نود تخصصی :9001-9007)
- cmd/agent-node/ (Guardian :9008)
- data/horizon-switch.db (لجر عملیاتی)
- docker-compose, frontend, config, static, templates
- start scripts, launchers, *.vbs

Genesis chain (root of trust): شاخه `main`

## Ports
Backend=8080 | Switch=8091 | Angel=9001-9007 | Agent=9008
