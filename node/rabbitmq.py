import pika
import json
import logging
import threading
from typing import Optional, Callable, Any, Dict, List
from datetime import datetime
from queue import Queue


class RabbitMQ:
    def __init__(self, host: str = 'localhost',
                 port: int = 5672,
                 username: str = 'admin',
                 password: str = 'password',
                 virtual_host: str = '/',
                 ):
        """Initialize RabbitMQ handler"""
        self.host = host
        self.port = port
        self.username = username
        self.password = password
        self.virtual_host = virtual_host
        self.connection: Optional[pika.BlockingConnection] = None
        self.channel: Optional[pika.channel.Channel] = None
        self.logger = logging.getLogger(__name__)
        self.consumer_threads: List[threading.Thread] = []
        self._is_running = False

    def connect(self) -> None:
        """Establish connection to RabbitMQ server"""
        try:
            credentials = pika.PlainCredentials(self.username, self.password)
            parameters = pika.ConnectionParameters(
                host=self.host,
                port=self.port,
                virtual_host=self.virtual_host,
                credentials=credentials,
                heartbeat=600
            )
            self.connection = pika.BlockingConnection(parameters)
            self.channel = self.connection.channel()
            self.logger.info("Successfully connected to RabbitMQ")
        except Exception as e:
            self.logger.error(f"Error connecting to RabbitMQ: {str(e)}")
            raise

    def exchange_declare(self, exchange_name : str, is_delay=False) -> None:
        """Declare an exchange with specified parameters"""
        try:
            exchange_type = 'x-delayed-message' if is_delay else 'direct'
            arguments = {'x-delayed-type': 'direct'} if is_delay else None
            self.channel.exchange_declare(
                exchange=exchange_name,
                exchange_type='x-delayed-message',
                durable=True,
                arguments=arguments
            )
        except Exception as e:
            self.logger.error(f"Error declaring exchange: {str(e)}")
            raise

    def setup_queue(self, queue_name: str, durable: bool = True, exchange_name='') -> None:
        """Declare a queue with specified parameters"""
        try:
            self.channel.queue_declare(
                queue=queue_name,
                durable=durable
            )
            if exchange_name:
                self.channel.queue_bind(
                    exchange=exchange_name,
                    queue=queue_name,
                    routing_key=queue_name
                )

        except Exception as e:
            self.logger.error(f"Error setting up queue: {str(e)}")
            raise

    def publish(self, queue_name: str, message: Any,
                persistent: bool = True, headers=None, exchange: str = '') -> bool:
        try:
            if isinstance(message, (dict, list)):
                message = json.dumps(message)
            if isinstance(message, str):
                message = message.encode()

            properties = pika.BasicProperties(
                delivery_mode=2 if persistent else 1,
                timestamp=int(datetime.now().timestamp()),
                content_type='application/json',
                headers=headers
            )

            self.channel.basic_publish(
                exchange=exchange,
                routing_key=queue_name,
                body=message,
                properties=properties
            )
            self.logger.info(f"Published message to queue: {queue_name}")
            return True
        except Exception as e:
            self.logger.error(f"Error publishing message: {str(e)}")
            return False

    def publish_batch(self, queue_name: str, messages: list) -> Dict[int, bool]:
        """Publish multiple messages to specified queue"""
        results = {}
        for i, message in enumerate(messages):
            results[i] = self.publish(queue_name, message)
        return results

    def process_message(self, ch, method, properties, body) -> None:
        """Default callback for processing messages"""
        try:
            message = body.decode()
            self.logger.info(f"Received message: {message}")
            ch.basic_ack(delivery_tag=method.delivery_tag)
        except Exception as e:
            self.logger.error(f"Error processing message: {str(e)}")
            ch.basic_nack(delivery_tag=method.delivery_tag, requeue=True)

    def _consume(self, queue_name: str, callback: Callable,
                 prefetch_count: int) -> None:
        """Internal consume method to run in thread"""
        try:
            # Create new connection and channel for this thread
            connection = pika.BlockingConnection(
                pika.ConnectionParameters(
                    host=self.host,
                    port=self.port,
                    virtual_host=self.virtual_host,
                    credentials=pika.PlainCredentials(self.username, self.password)
                )
            )
            channel = connection.channel()
            channel.basic_qos(prefetch_count=prefetch_count)

            channel.basic_consume(
                queue=queue_name,
                on_message_callback=callback,
                auto_ack=True
            )

            self.logger.info(f"Started consuming from queue: {queue_name}")

            while self._is_running:
                try:
                    channel.connection.process_data_events(time_limit=1)  # Non-blocking
                except Exception as e:
                    self.logger.error(f"Error processing events: {str(e)}")
                    break

            channel.close()
            connection.close()

        except Exception as e:
            self.logger.error(f"Error in consumer thread: {str(e)}")

    def start_consuming(self, queue_name: str,
                        callback: Optional[Callable] = None,
                        prefetch_count: int = 1,
                        num_threads: int = 1) -> None:
        """Start consuming messages from the specified queue using threads"""
        try:
            self._is_running = True
            message_callback = callback if callback else self.process_message

            # Create consumer threads
            for _ in range(num_threads):
                consumer_thread = threading.Thread(
                    target=self._consume,
                    args=(queue_name, message_callback, prefetch_count)
                )
                consumer_thread.daemon = True  # Thread will exit when main thread exits
                consumer_thread.start()
                self.consumer_threads.append(consumer_thread)

            self.logger.info(f"Started {num_threads} consumer threads for queue: {queue_name}")

        except Exception as e:
            self.logger.error(f"Error starting consumers: {str(e)}")
            self.stop()

    def stop(self) -> None:
        """Stop all consumers and close connections"""
        try:
            self._is_running = False

            # Wait for consumer threads to finish
            for thread in self.consumer_threads:
                thread.join(timeout=5.0)

            if self.channel:
                self.channel.close()
            if self.connection:
                self.connection.close()

            self.logger.info("Successfully closed RabbitMQ connection")

        except Exception as e:
            self.logger.error(f"Error closing connection: {str(e)}")


def example_usage():
    # Configure logging
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )

    # Create RabbitMQ instance
    rabbit = RabbitMQ(
        host='localhost',
        username='admin',
        password='password'
    )
    queue = Queue()

    try:
        # Connect to RabbitMQ
        rabbit.connect()
        queue_name = 'test_queue'

        # Setup queue
        rabbit.setup_queue(queue_name)
        for i in range(10):
            rabbit.publish(queue_name, {'message': 'Hello, RabbitMQ!'})

        # Custom callback
        def custom_callback(ch, method, properties, body):
            message = json.loads(body.decode())
            print(f"Custom processing: {message}")
            ch.basic_ack(delivery_tag=method.delivery_tag)
            queue.put(message)

        # Start consuming with multiple threads
        rabbit.start_consuming(
            queue_name,
            callback=custom_callback,
            num_threads=3  # Start 3 consumer threads
        )

        # Keep main thread alive
        try:
            while True:
                if queue.qsize() == 10:
                    print("All messages received")

                    break
                # You can do other work here
                # The consumer threads will run in the background
        except KeyboardInterrupt:
            print("Stopping...")

    except Exception as e:
        logging.error(f"Error in main: {str(e)}")
    finally:
        rabbit.stop()


if __name__ == "__main__":
    example_usage()
