from rabbitmq import RabbitMQ
from config import RABBITMQ_HOST, RABBITMQ_PORT, RABBITMQ_USER, RABBITMQ_PASS, TASK_QUEUE, RESULT_QUEUE, DELAY_EXCHANGE
from task_process import TaskProcessor
from queue import Queue
from threading import Thread
import json
import time
import logging
from model.task import TaskResult
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')

rabbitmq = RabbitMQ(
    host=RABBITMQ_HOST,
    port=RABBITMQ_PORT,
    username=RABBITMQ_USER,
    password=RABBITMQ_PASS
)
task_queue = Queue()
result_queue = Queue()
task_processor = TaskProcessor(task_queue, result_queue)


def publish_results(queue_name: str,):
    retry_count = {}  # Dictionary to keep track of retry counts for each task

    while True:
        if result_queue.empty():
            time.sleep(1)
            continue

        result: TaskResult = result_queue.get()
        logging.info(f"Publishing task result: {result}")

        # Initialize retry count for this task if not already tracked
        if result.task_id not in retry_count:
            retry_count[result.task_id] = 0

        succeed = rabbitmq.publish(queue_name, str(result.to_json()))
        logging.info(f"Published task result: {result}")

        if not succeed:
            retry_count[result.task_id] += 1
            if retry_count[result.task_id] < 3:
                result_queue.put(result)
                time.sleep(1)
            else:
                logging.error(f"Failed to publish task {result.task_id} after 3 retries. Giving up.")
                del retry_count[result.task_id]  # Remove the task from retry tracking
        else:
            del retry_count[result.task_id]  # Remove the task from retry tracking if successful


def main():
    rabbitmq.connect()

    def callback(ch, method, properties, body):
        message = json.loads(body.decode())
        task_queue.put(message)
    rabbitmq.start_consuming(TASK_QUEUE, callback=callback, num_threads=3, prefetch_count=1)

    thread = Thread(target=task_processor.start)
    thread2 = Thread(target=publish_results, args=(RESULT_QUEUE,))
    thread.start()
    thread2.start()


if __name__ == '__main__':
    main()
