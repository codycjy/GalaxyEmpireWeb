from rabbitmq import RabbitMQ
from config import RABBITMQ_HOST, RABBITMQ_PORT, RABBITMQ_USER, RABBITMQ_PASS, TASK_QUEUE, DELAY_EXCHANGE
from model.task import Task
from model.task import TaskType
from model.target import Target
from model.user import Account
from model.fleet import Fleet
import uuid
rabbit = RabbitMQ(
    host=RABBITMQ_HOST,
    port=RABBITMQ_PORT,
    username=RABBITMQ_USER,
    password=RABBITMQ_PASS

)
rabbit.connect()
queue_name = TASK_QUEUE

# Setup queue
rabbit.setup_queue(queue_name, exchange_name=DELAY_EXCHANGE)
t = Task(0,str(uuid.uuid4()),TaskType.LOGIN, Account("saltfish", "A123456", "g26", "g26"), Fleet(), 0, target=Target(1, 1, 1),)
json_task=t.to_json()
print(json_task)


for i in range(1):
    rabbit.publish(queue_name,
                   message=json_task,
                   headers={'x-delay': 1000},
                   exchange=DELAY_EXCHANGE
                   )
    print(' [x] Sent message')
