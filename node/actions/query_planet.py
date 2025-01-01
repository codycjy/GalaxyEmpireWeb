import logging
import json
from queue import Queue
from model.task import Task, TaskStatus, TaskResult
from galaxy_core import Galaxy

logger = logging.getLogger(__name__)


def query_planet_action(task: Task, result_queue: Queue):
    task_status = TaskStatus.SUCCESS
    uuid = task.uuid
    back_ts = -1
    task_result = TaskResult(task_id=task.task_id, status=task_status, task_type=task.task_type, back_ts=back_ts, uuid=uuid, msg="", err_msg="")
    try:
        galaxy = Galaxy(task.account, result_queue=result_queue)
        login_response = galaxy.login()
        if login_response.status != 0:
            logger.warning(f"Login failed: {login_response.err_msg}")
            task_status = TaskStatus.FAILED
            task_result.err_msg = login_response.err_msg
            raise Exception(login_response.err_msg)
        query_response = galaxy.query_planet_id(task)
        if query_response.status != 0:
            task_status = TaskStatus.FAILED
            task_result.err_msg = query_response.err_msg
        task_result.msg = json.dumps(query_response.data)
    except Exception as e:
        task_status = TaskStatus.FAILED
        task_result.err_msg = str(e)
    finally:
        result_queue.put(task_result)
