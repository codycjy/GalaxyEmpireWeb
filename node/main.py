import logging
import signal
import sys
import threading
import time
from queue import Queue, Empty
from threading import Thread, Event
import json

from model.task import TaskResult
from rabbitmq import RabbitMQPublisher, RabbitMQConsumer
from task_process import TaskProcessor
from config import (
    RABBITMQ_HOST, RABBITMQ_PORT, RABBITMQ_USER, RABBITMQ_PASS,
    TASK_QUEUE, RESULT_QUEUE, DELAY_EXCHANGE
)

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(lineno)d - %(message)s'
)
logger = logging.getLogger(__name__)


class Worker:
    def __init__(self):
        self.task_queue = Queue()
        self.result_queue = Queue()
        self.shutdown_event = Event()
        self.threads = []

        # Initialize Publisher and Consumer
        self.publisher = RabbitMQPublisher(
            host=RABBITMQ_HOST,
            port=RABBITMQ_PORT,
            username=RABBITMQ_USER,
            password=RABBITMQ_PASS
        )
        self.consumer = RabbitMQConsumer(
            host=RABBITMQ_HOST,
            port=RABBITMQ_PORT,
            username=RABBITMQ_USER,
            password=RABBITMQ_PASS
        )
        self.task_processor = TaskProcessor(self.task_queue, self.result_queue)

    def publish_results(self, queue_name: str):
        """Thread target for publishing results to RabbitMQ."""
        logger.info("Result publisher thread started")
        while not self.shutdown_event.is_set():
            try:
                result: TaskResult = self.result_queue.get(timeout=1)
                retry_count = 0
                max_retries = 3
                backoff = 1

                while retry_count < max_retries and not self.shutdown_event.is_set():
                    try:
                        message = result.to_dict()
                        message['status'] = message['status'].value
                        message['task_type'] = message['task_type'].value
                        success = self.publisher.publish(queue_name, message)
                        if success:
                            logger.info(f"Published result for task {result.task_id}")
                            break
                        else:
                            raise Exception("Publish returned False")
                    except Exception as e:
                        retry_count += 1
                        logger.error(f"Publish error: {e}, retry {retry_count}/{max_retries}")
                        time.sleep(backoff)
                        backoff *= 2  # Exponential backoff

                if retry_count >= max_retries:
                    logger.error(f"Failed to publish task {result.task_id} after {max_retries} retries")
                    # Optionally, push to a dead-letter queue or handle accordingly

            except Empty:
                continue
            except Exception as e:
                logger.exception(f"Error in publish_results: {e}")
                time.sleep(1)
        logger.info("Result publisher thread exiting")

    def handle_consumed_message(self, ch, method, properties, body):
        """Callback for consumed messages."""
        try:
            message = json.loads(body.decode())
            self.task_queue.put(message)
            ch.basic_ack(delivery_tag=method.delivery_tag)
            logger.info(f"Received and acknowledged message: {message}")
        except json.JSONDecodeError:
            logger.error(f"Invalid JSON message: {body}")
            ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)
        except Exception as e:
            logger.exception(f"Error processing message: {e}")
            ch.basic_nack(delivery_tag=method.delivery_tag, requeue=True)

    def consume_messages(self):
        """Start consuming messages using RabbitMQConsumer."""
        logger.info("Starting message consumer")
        self.consumer.start_consuming(TASK_QUEUE, self.handle_consumed_message)

    def start(self):
        logger.info("Starting worker...")

        # Start Task Processor in a separate thread
        processor_thread = Thread(target=self.task_processor.start, daemon=True)
        self.threads.append(processor_thread)
        processor_thread.start()

        # Start Result Publisher in a separate thread
        publisher_thread = Thread(target=self.publish_results, args=(RESULT_QUEUE,), daemon=True)
        self.threads.append(publisher_thread)
        publisher_thread.start()

        # Start Consumer
        consumer_thread = Thread(target=self.consume_messages, daemon=True)
        self.threads.append(consumer_thread)
        consumer_thread.start()

    def shutdown(self):
        logger.info("Initiating shutdown...")

        self.shutdown_event.set()

        # Stop Consumer
        if self.consumer:
            try:
                self.consumer.stop_consuming()
            except Exception as e:
                logger.error(f"Error stopping RabbitMQ consumer: {e}")

        # Stop Publisher
        if self.publisher:
            try:
                self.publisher.stop()
            except Exception as e:
                logger.error(f"Error stopping RabbitMQ publisher: {e}")

        # Stop Task Processor
        if self.task_processor:
            try:
                self.task_processor.shutdown()
            except Exception as e:
                logger.error(f"Error shutting down task processor: {e}")

        # Wait for all threads to finish
        for thread in self.threads:
            try:
                thread.join(timeout=5)
            except Exception as e:
                logger.error(f"Error joining thread: {e}")

        logger.info("Shutdown complete")


def run_forever():
    worker = Worker()

    def signal_handler(signum, frame):
        logger.info(f"Received signal {signum}, shutting down...")
        worker.shutdown()
        sys.exit(0)

    # Register signal handlers
    signal.signal(signal.SIGTERM, signal_handler)
    signal.signal(signal.SIGINT, signal_handler)

    try:
        worker.start()
        # Keep the main thread alive to handle signals
        while not worker.shutdown_event.is_set():
            time.sleep(1)
    except Exception as e:
        logger.exception(f"Worker crashed: {e}")
        worker.shutdown()
        time.sleep(5)
        sys.exit(1)


if __name__ == '__main__':
    run_forever()
