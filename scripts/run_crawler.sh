#!/bin/bash
# Run crawler in detached mode for proper Playwright stability
cd /home/datdt/tikiclone
exec node scripts/tiki_crawler.mjs --target=${TARGET_PRODUCTS:-50000} --skip-images 2>&1