import pika
import json

# 连接配置
credentials = pika.PlainCredentials('admin', 'password')
connection = pika.BlockingConnection(
    pika.ConnectionParameters(host='localhost', credentials=credentials)

)
channel = connection.channel()

# 声明延时交换机
# 注意：exchange_type 需要设置为 x-delayed-message
channel.exchange_declare(
    exchange='delayed_exchange',
    exchange_type='x-delayed-message',
    arguments={'x-delayed-type': 'direct'}  # 可以是 direct、topic、fanout
)

# 声明队列
channel.queue_declare(queue='delayed_queue', durable=True)
channel.queue_bind(
    queue='delayed_queue',
    exchange='delayed_exchange',
    routing_key='delayed_routing_key'
)

# 发送延时消息
message = "Hello Delayed Message!"
delay_time = 10000  # 延时时间(毫秒)

# 使用 headers 设置延时时间
channel.basic_publish(
    exchange='delayed_exchange',
    routing_key='delayed_routing_key',
    body=json.dumps(message),
    properties=pika.BasicProperties(
        delivery_mode=2,  # 消息持久化
        headers={'x-delay': delay_time}  # 设置延时时间
    ),
)

print(f" [x] Sent delayed message: {message}")
connection.close()
