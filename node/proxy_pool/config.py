import os
import sys
import logging

if os.getenv("PROXY", ""):
    print("************ PROXY ENABLED ************")
    API_URL = os.getenv("PROXY_API_URL")
    if not API_URL:
        logging.error("PROXY_API_URL not set")
        sys.exit(1)
