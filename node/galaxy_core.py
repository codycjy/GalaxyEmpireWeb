import time
from queue import Queue
from network import Network
from model.user import Account
from rabbitmq import RabbitMQ
from model.task import Task, TaskType, MissionType
from model.fleet import Fleet
from model.target import Target
# from config import fleet_config, ShipToID


class Galaxy(Network):
    def __init__(self, user: Account, result_queue: Queue):
        super().__init__(user)
        self.user = user
        self.rabbitmq = RabbitMQ()
        self.result_queue = result_queue

    def prepare_fleet(self, task: Task):
        """
        Prepare fleet for the task

        Args:
            task (Task): Task object

        Returns:
            dict: Dictionary containing the arguments and token

        """
        PREPARE_FLEET_END_POINT = "game.php?page=my_fleet1"
        args = {}
        if task.task_type == TaskType.ATTACK:
            args['mission'] = MissionType.ATTACK.value
            args['type'] = TaskType.ATTACK.value
            args['galaxy'] = task.target.galaxy
            args['system'] = task.target.system
            args['planet'] = task.target.planet
            args['speed'] = 10
            fleet = task.fleet.to_fleet()
            args.update(fleet)
            response = self._post(PREPARE_FLEET_END_POINT, args)
        elif task.task_type == TaskType.EXPLORE:
            args['mission'] = MissionType.EXPLORE.value
            args['type'] = TaskType.EXPLORE.value
            args['galaxy'] = task.target.galaxy
            args['system'] = task.target.system
            args['planet'] = task.target.planet
            args['speed'] = 10
            fleet = task.fleet.to_fleet()
            args.update(fleet)
            response = self._post(PREPARE_FLEET_END_POINT, args)
        if response['status'] == 0:
            token = response['data']['result']['token']
            print(token)
            print("Fleet prepared")
            return {'args': args, 'token': token}

    def handle_attack_task(self, task):
        total_finish_ts = -1
        for _ in range(task.repeat):
            finish_ts = self.handle_single_attack_task(task)
            if finish_ts == -1:
                print("Error sending fleet")
                # TODO: stats update
            total_finish_ts = max(total_finish_ts, finish_ts)
            time.sleep(1)
        return total_finish_ts

    def handle_single_attack_task(self, task: Task) -> int:
        """
        Handle single attack Task

        Args:
            task (Task): Task object

        Returns:
            int: backtime of the fleet, if successful, -1 otherwise



            """
        result = self.prepare_fleet(task)
        if not result:
            print("Error preparing fleet")
            return -1

        args = result['args']
        token = result['token']
        SEND_FLEET_END_POINT = "game.php?page=fleet3"
        args['token'] = token
        response = self._post(SEND_FLEET_END_POINT, args)
        if response['status'] != 0:
            print("Error sending fleet")
            print(response)
            return -1
        print("Fleet sent")
        print(response)
        backtime: int = response['data']['result']['back_ts']
        return backtime

    def handle_explore_task(self, task) -> int:
        """
        Handle explore Task

        Args:
            task (Task): Task object

        Returns:
            int: backtime of the fleet, if successful, -1 otherwise
        """
        total_finish_ts = -1
        for _ in range(task.repeat):
            result = self.prepare_fleet(task)
            if not result:
                print("Error preparing fleet")
                return -1

            args = result['args']
            token = result['token']
            END_POINT = "game.php?page=fleet3"
            args['token'] = token
            args['staytime'] = 1
            response = self._post(END_POINT, args)
            if response['status'] != 0:
                print("Error sending fleet")
                print(response)
                continue
            print("Fleet sent")
            print(response)
            backtime: int = response['data']['result']['back_ts']
            total_finish_ts = max(total_finish_ts, backtime)
            time.sleep(1)
        return total_finish_ts

    def handle_escape_task(self, task):  # TODO:
        # 处理逃跑任务
        pass


if __name__ == '__main__':
    pass
