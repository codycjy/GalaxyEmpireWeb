import logging
from queue import Queue
from model.task import Task, TaskType, TaskResult, TaskStatus
from galaxy_core import Galaxy

logger = logging.getLogger(__name__)


def attack_action(task: Task, result_queue: Queue):
    """
    Handle attack task and put result in queue.

    Args:
        task (Task): The attack task to process
        result_queue (Queue): Queue to put task results
    """
    task_status = TaskStatus.SUCCESS
    uuid = task.uuid
    back_ts = -1

    try:
        if task.task_type != TaskType.ATTACK:
            logger.error(f"Invalid task type for attack_action: {task.task_type}")
            task_status = TaskStatus.FAILED
        else:
            logger.info(f"Processing attack task {task.task_id}")
            galaxy = Galaxy(task.account, result_queue=result_queue)
            response = galaxy.handle_attack_task(task)

            if response.status != 0:
                logger.warning(f"Attack task failed: {response.err_msg}")
                task_status = TaskStatus.FAILED
            else:
                back_ts = response.data.get('total_finish_ts', -1)
                if back_ts == -1:
                    logger.warning("Invalid back_ts in response")
                    task_status = TaskStatus.FAILED
                else:
                    logger.info(f"Attack task {task.task_id} completed successfully")

    except Exception as e:
        logger.error(f"Error processing attack task {task.task_id}: {str(e)}", exc_info=True)
        task_status = TaskStatus.FAILED
    finally:
        result = TaskResult(
            task_id=task.task_id,
            status=task_status,
            task_type=task.task_type,
            back_ts=back_ts,
            uuid=uuid
        )
        result_queue.put(result)
        logger.debug(f"Result queued for task {task.task_id}: {result}")


def explore_action(task: Task, result_queue: Queue):
    """
    Handle explore task and put result in queue.

    Args:
        task (Task): The explore task to process
        result_queue (Queue): Queue to put task results
    """
    task_status = TaskStatus.SUCCESS
    back_ts = -1

    try:
        if task.task_type != TaskType.EXPLORE:
            logger.error(f"Invalid task type for explore_action: {task.task_type}")
            task_status = TaskStatus.FAILED
        else:
            logger.info(f"Processing explore task {task.task_id}")
            galaxy = Galaxy(task.account, result_queue)
            response = galaxy.handle_explore_task(task)

            if response.status != 0:
                logger.warning(f"Explore task failed: {response.err_msg}")
                task_status = TaskStatus.FAILED
            else:
                back_ts = response.data.get('total_finish_ts', -1)
                if back_ts == -1:
                    logger.warning("Invalid back_ts in response")
                    task_status = TaskStatus.FAILED
                else:
                    logger.info(f"Explore task {task.task_id} completed successfully")

    except Exception as e:
        logger.error(f"Error processing explore task {task.task_id}: {str(e)}", exc_info=True)
        task_status = TaskStatus.FAILED
    finally:
        result = TaskResult(
            task_id=task.task_id,
            status=task_status,
            task_type=task.task_type,
            back_ts=back_ts,
            uuid=task.uuid
        )
        result_queue.put(result)
        logger.debug(f"Result queued for task {task.task_id}: {result}")
