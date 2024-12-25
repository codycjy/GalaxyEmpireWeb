import os
RABBITMQ_HOST = os.environ.get('RABBITMQ_HOST', "localhost")
RABBITMQ_USER = os.environ.get('RABBITMQ_USER', 'admin')
RABBITMQ_PASS = os.environ.get('RABBITMQ_PASS', 'password')
TASK_QUEUE = os.environ.get('TASK_QUEUE', 'task_queue')
RESULT_QUEUE = os.environ.get('RESULT_QUEUE', 'result_queue')
DELAY_EXCHANGE = os.environ.get('DELAY_EXCHANGE', 'delayed_exchange')
PROXY_BASE_URL = os.environ.get('PROXY_ENDPOINT', 'http://localhost:5010')
PROXY_AUTH_USER = os.environ.get('PROXY_AUTH_USER', 'user')
PROXY_AUTH_PASS = os.environ.get('PROXY_AUTH_PASS', 'password')
if os.getenv('PROXY', False):
    print("************ PROXY ENABLED ************")

port = os.environ.get('RABBITMQ_PORT', "5672")
try:
    RABBITMQ_PORT = int(port)
except:
    RABBITMQ_PORT = 5672

serverUrlList = {  # TODO: Get Server URL from the server
    'g26': 'http://45.33.62.217/g26/',
    'ze': 'http://45.33.39.137/zadc/',
}

ShipToID = {
    'ds': 'ship214',
    'de': 'ship213',
    'cargo': 'ship203',
    'bs': 'ship207',
    'satellite': 'ship210',
    'lf': 'ship204',
    'hf': 'ship205',
    'cr': 'ship206',
    'dr': 'ship215',
    'bomb': 'ship211',
    'guard': 'ship216'
}
