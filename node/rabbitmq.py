import pika
import json
import logging
import threading
import time
import uuid
from typing import Callable, Optional


class RabbitMQPublisher:
    """RabbitMQ Publisher with thread-safe publish method and reconnection logic."""

    def __init__(self, host: str = 'localhost',
                 port: int = 5672,
                 username: str = 'guest',
                 password: str = 'guest',
                 virtual_host: str = '/',
                 heartbeat: int = 600):
        self.host = host
        self.port = port
        self.username = username
        self.password = password
        self.virtual_host = virtual_host
        self.heartbeat = heartbeat

        self.connection: Optional[pika.BlockingConnection] = None
        self.channel: Optional[pika.channel.Channel] = None

        self._publishing_lock = threading.Lock()
        self._stop_event = threading.Event()
        self._reconnect_delay = 1
        self._max_reconnect_delay = 30

        self._logger = logging.getLogger(__name__)
        self.connect()

    def connect(self):
        """Establish connection and channel for publishing."""
        with self._publishing_lock:
            while not self._stop_event.is_set():
                try:
                    credentials = pika.PlainCredentials(self.username, self.password)
                    parameters = pika.ConnectionParameters(
                        host=self.host,
                        port=self.port,
                        virtual_host=self.virtual_host,
                        credentials=credentials,
                        heartbeat=self.heartbeat,
                        blocked_connection_timeout=300
                    )
                    self.connection = pika.BlockingConnection(parameters)
                    self.channel = self.connection.channel()
                    self._logger.info("Publisher connected to RabbitMQ")
                    self._reconnect_delay = 1  # Reset reconnect delay after successful connection
                    break
                except pika.exceptions.AMQPConnectionError as e:
                    self._logger.error(f"Publisher connection failed: {e}. Reconnecting in {self._reconnect_delay}s...")
                    time.sleep(self._reconnect_delay)
                    self._reconnect_delay = min(self._reconnect_delay * 2, self._max_reconnect_delay)

    def publish(self, queue_name: str, message: dict, persistent: bool = True) -> bool:
        """Publish a message to the specified queue."""
        with self._publishing_lock:
            if self.connection is None or self.connection.is_closed:
                self._logger.warning("Publisher connection closed, attempting to reconnect...")
                self.connect()
            try:
                properties = pika.BasicProperties(
                    delivery_mode=2 if persistent else 1,
                    content_type='application/json'
                )
                self.channel.basic_publish(
                    exchange='',
                    routing_key=queue_name,
                    body=json.dumps(message),
                    properties=properties
                )
                self._logger.debug(f"Published message to {queue_name}: {message}")
                return True
            except (pika.exceptions.ConnectionClosed,
                    pika.exceptions.ChannelClosed,
                    pika.exceptions.StreamLostError) as e:
                self._logger.error(f"Publish failed due to connection error: {e}")
                self.close()
                return False
            except Exception as e:
                self._logger.exception(f"Unexpected error during publish: {e}")
                return False

    def close(self):
        """Close publisher connection and channel."""
        with self._publishing_lock:
            try:
                if self.channel and not self.channel.is_closed:
                    self.channel.close()
                    self._logger.info("Publisher channel closed")
            except Exception as e:
                self._logger.error(f"Error closing publisher channel: {e}")
            try:
                if self.connection and not self.connection.is_closed:
                    self.connection.close()
                    self._logger.info("Publisher connection closed")
            except Exception as e:
                self._logger.error(f"Error closing publisher connection: {e}")

    def stop(self):
        """Stop the publisher and close connections."""
        self._stop_event.set()
        self.close()


class RabbitMQConsumer:
    """RabbitMQ Consumer running in its own thread with reconnection logic."""

    def __init__(self, host: str = 'localhost',
                 port: int = 5672,
                 username: str = 'guest',
                 password: str = 'guest',
                 virtual_host: str = '/',
                 heartbeat: int = 600,
                 prefetch_count: int = 1):
        self.host = host
        self.port = port
        self.username = username
        self.password = password
        self.virtual_host = virtual_host
        self.heartbeat = heartbeat
        self.prefetch_count = prefetch_count

        self.connection: Optional[pika.BlockingConnection] = None
        self.channel: Optional[pika.channel.Channel] = None
        self.consumer_tag: Optional[str] = None

        self._stop_event = threading.Event()
        self._consumer_thread = None
        self._reconnect_delay = 1
        self._max_reconnect_delay = 30

        self._logger = logging.getLogger(__name__)

    def start_consuming(self, queue_name: str, callback: Callable):
        """Start the consumer thread."""
        if self._consumer_thread and self._consumer_thread.is_alive():
            self._logger.warning("Consumer is already running")
            return
        self._consumer_thread = threading.Thread(
            target=self._consumer_loop, args=(queue_name, callback), daemon=True
        )
        self._consumer_thread.start()
        self._logger.info("Consumer thread started")

    def _consumer_loop(self, queue_name: str, callback: Callable):
        """Consumer loop that handles connection, consuming, and reconnection."""
        while not self._stop_event.is_set():
            try:
                self._establish_connection()
                self.channel.basic_qos(prefetch_count=self.prefetch_count)
                self.consumer_tag = self.channel.basic_consume(
                    queue=queue_name,
                    on_message_callback=callback,
                    auto_ack=False
                )
                self._logger.info(f"Started consuming on queue: {queue_name}")
                self.channel.start_consuming()
            except pika.exceptions.AMQPConnectionError as e:
                self._logger.error(f"Consumer connection error: {e}")
            except pika.exceptions.ChannelClosedByBroker as e:
                self._logger.error(f"Consumer channel closed by broker: {e}")
            except Exception as e:
                self._logger.exception(f"Unexpected consumer error: {e}")
            finally:
                if self.connection and not self.connection.is_closed:
                    try:
                        self.connection.close()
                        self._logger.info("Consumer connection closed")
                    except Exception as e:
                        self._logger.error(f"Error closing consumer connection: {e}")
                if not self._stop_event.is_set():
                    self._logger.info(f"Reconnecting consumer after {self._reconnect_delay}s...")
                    time.sleep(self._reconnect_delay)
                    self._reconnect_delay = min(self._reconnect_delay * 2, self._max_reconnect_delay)

    def _establish_connection(self):
        """Establish connection and channel for consumer."""
        credentials = pika.PlainCredentials(self.username, self.password)
        parameters = pika.ConnectionParameters(
            host=self.host,
            port=self.port,
            virtual_host=self.virtual_host,
            credentials=credentials,
            heartbeat=self.heartbeat,
            blocked_connection_timeout=300
        )
        self.connection = pika.BlockingConnection(parameters)
        self.channel = self.connection.channel()
        self.channel.basic_qos(prefetch_count=self.prefetch_count)
        self._logger.info("Consumer connected to RabbitMQ")
        self._reconnect_delay = 1  # Reset reconnect delay after successful connection

    def stop_consuming(self):
        """Stop consuming and close connections."""
        self._stop_event.set()
        if self.channel and self.channel.is_open:
            try:
                self.channel.stop_consuming()
                self._logger.info("Consumer stopping...")
            except Exception as e:
                self._logger.error(f"Error stopping consumer: {e}")
        if self._consumer_thread:
            self._consumer_thread.join(timeout=5)
            self._logger.info("Consumer thread stopped")

        try:
            if self.connection and not self.connection.is_closed:
                self.connection.close()
                self._logger.info("Consumer connection closed")
        except Exception as e:
            self._logger.error(f"Error closing consumer connection: {e}")
