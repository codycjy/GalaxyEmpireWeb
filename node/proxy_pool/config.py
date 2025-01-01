import os
import sys
import logging

API_URL = os.getenv('PROXY_API_URL')
if not API_URL:
    logging.error("PROXY_API_URL not set")
    sys.exit(1)
