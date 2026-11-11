#!/usr/bin/env bash
set -u
C=$'\e[36m'; G=$'\e[32m'; R=$'\e[31m'; X=$'\e[0m'

case "${1:-}" in
  up)
    echo "${C}درحال ساخت ایمیج و راه‌اندازی کل شبکه...${X}"
    docker compose up -d --build
    ;;
  down)
    echo "${C}در حال خاموش کردن شبکه...${X}"
    docker compose down
    ;;
  restart)
    docker compose restart
    ;;
  status)
    docker compose ps
    ;;
  logs)
    docker compose logs -f "${2:-}"
    ;;
  test)
    echo "${C}=== تست سلامت همه سرویس‌ها ===${X}"
    for p in 8080 8091 8092 9001 9002 9003 9004 9005 9006 9007 9008 9300; do
      echo -n "  پورت :$p  "
      if curl -s --max-time 2 "http://127.0.0.1:$p/health" > /dev/null; then
        echo "${G}✅ OK${X}"
      else
        echo "${R}❌ FAIL${X}"
      fi
    done
    ;;
  *)
    echo "استفاده: ./horizon-docker.sh [up|down|restart|status|logs|test]"
    ;;
esac
