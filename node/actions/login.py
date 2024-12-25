import logging
from queue import Queue
from network import Network
from model.task import TaskType, TaskResult, TaskStatus, Task

logger = logging.getLogger(__name__)


def login_action(task: Task, result_queue: Queue):
    """
    Handle login task and put result in queue.

    Args:
        task (Task): The login task to process
        result_queue (Queue): Queue to put task results
    """
    try:
        if task.task_type != TaskType.LOGIN:
            logger.error(f"Invalid task type for login_action: {task.task_type}")
            status = TaskStatus.FAILED
        else:
            logger.info(f"Processing login task for user: {task.account.username}")
            network = Network(task.account)
            response = network.login()

            status = TaskStatus.SUCCESS if response.status == 0 else TaskStatus.FAILED

            if status == TaskStatus.SUCCESS:
                logger.info("Login successful")
            else:
                logger.warning(f"Login failed: {response.err_msg}")

    except Exception as e:
        logger.error(f"Error during login: {str(e)}", exc_info=True)
        status = TaskStatus.FAILED
    finally:
        result = TaskResult(
            task_id=0,
            status=status,
            task_type=TaskType.LOGIN,
            uuid=task.uuid
        )
        result_queue.put(result)
        logger.debug(f"Login result queued: {result}")
