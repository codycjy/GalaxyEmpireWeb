from queue import Queue
from model.task import Task, TaskType, TaskResult
from galaxy_core import Galaxy


def attack_action(task: Task, result_queue: Queue):
    task_status = 0
    back_ts = -1
    try:
        if task.task_type == TaskType.ATTACK:
            print("Attack task")
            back_ts = Galaxy(task.account, result_queue=result_queue).handle_attack_task(task)
            if back_ts == -1:
                print("Error in attack task")
                task_status = -1
            else:
                print("Attack task success")
    except Exception as e:
        print(f"Error processing task: {e}")
        return
    result = TaskResult(task_id=task.task_id, status=task_status, task_type=task.task_type, back_ts=back_ts)
    result_queue.put(result)


def explore_action(task: Task, result_queue: Queue):
    task_status = 0
    back_ts = -1
    try:
        if task.task_type == TaskType.EXPLORE:
            print("Explore task")
            back_ts = Galaxy(task.account, result_queue).handle_explore_task(task)
            if back_ts == -1:
                print("Error in explore task")
                task_status = -1
            else:
                print("Explore task success")
    except Exception as e:
        print(f"Error processing task: {e}")
        return
    result_queue.put(TaskResult(task_id=task.task_id, status=task_status, task_type=task.task_type, back_ts=back_ts))
