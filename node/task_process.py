import logging
from queue import Queue, Empty
from concurrent.futures import ThreadPoolExecutor
from model.task import Task, TaskType, TaskResult, TaskStatus
from actions.login import login_action
from actions.attack import attack_action
from actions.attack import explore_action


class TaskProcessor:
    def __init__(self, task_queue: Queue, result_queue: Queue, max_workers=5):
        self.task_queue = task_queue
        self.result_queue = result_queue
        self.executor = ThreadPoolExecutor(max_workers=max_workers)
        self.is_running = True
        self._logger = logging.getLogger(__name__)

    def _process(self, task_dict: dict) -> None:
        try:
            # Parse the task dictionary into Task object
            task = Task.from_dict(task_dict)
            self._logger.info(f"Processing task: {task.uuid} - {task.task_id} - Type: {task.task_type}")

            action_map = {
                TaskType.LOGIN: login_action,
                TaskType.ATTACK: attack_action,
                TaskType.EXPLORE: explore_action
            }

            action = action_map.get(task.task_type)
            if action:
                self.executor.submit(
                    action,
                    task,
                    self.result_queue
                )
            else:
                self._logger.error(f"Unknown task type: {task.task_type}")
                error_result = TaskResult(
                    task_id=task.task_id,
                    status=TaskStatus.FAILED,
                    task_type=task.task_type,
                    uuid=task.uuid
                )
                self.result_queue.put(error_result)

        except Exception as e:
            self._logger.exception(f"Error processing task: {str(e)}")
            # If we can't parse the task, try to get task_id from dict
            task_id = task_dict.get('task_id', 0)
            task_type_value = task_dict.get('task_type', TaskType.LOGIN.name)
            try:
                task_type = TaskType(task_type_value) if isinstance(task_type_value, str) else TaskType(task_type_value)
            except ValueError:
                task_type = TaskType.LOGIN  # Default or handle appropriately
            uuid = task_dict.get('uuid', '')

            error_result = TaskResult(
                task_id=task_id,
                status=TaskStatus.FAILED,
                task_type=task_type,
                uuid=uuid
            )
            self.result_queue.put(error_result)
        finally:
            self.task_queue.task_done()

    def start(self):
        self._logger.info("TaskProcessor started")
        try:
            while self.is_running:
                try:
                    # Add timeout to make it interruptible
                    task_dict = self.task_queue.get(timeout=1)
                    self._process(task_dict)
                except Empty:
                    continue
        except Exception as e:
            self._logger.exception(f"Fatal error in task processor: {e}")
        finally:
            self.shutdown()

    def shutdown(self):
        self.is_running = False
        self.executor.shutdown(wait=True)
        self._logger.info("TaskProcessor shutdown complete")
