import logging
from queue import Queue
from threading import Thread
from model.task import Task, TaskType, TaskResult
from actions.login import login_action
from actions.attack import attack_action, explore_action


class TaskProcessor:
    def __init__(self, task_queue: Queue, result_queue: Queue):
        self.task_queue = task_queue
        self.result_queue = result_queue

    def _process(self, task):
        task_status = 0
        try:
            logging.info(f"Processing task: {task}")
            task = Task.from_dict(task)  # pyright: ignore
        except Exception as e:
            print(f"Error parsing task: {e}")
            return
        task_id = task.task_id
        try:
            if task.task_type == TaskType.LOGIN:
                Thread(target=login_action, args=(task, self.result_queue)).start()
            elif task.task_type == TaskType.ATTACK:
                Thread(target=attack_action, args=(task, self.result_queue)).start()
            elif task.task_type == TaskType.EXPLORE:
                Thread(target=explore_action, args=(task, self.result_queue)).start()
        except Exception as e:
            print(f"Error processing task: {e}")
            return

    def start(self):
        while True:
            task = self.task_queue.get()
            self._process(task)
            # self.task_queue.task_done()


if __name__ == '__main__':
    task_queue = Queue()
    result_queue = Queue()
    task_processor = TaskProcessor(task_queue, result_queue)
    task_processor.start()
