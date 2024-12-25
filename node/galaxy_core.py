import logging
import time
from queue import Queue
from network import Network, NetworkResponse
from model.user import Account
from model.task import Task, TaskType, MissionType
from model.fleet import Fleet
from model.target import Target

logger = logging.getLogger(__name__)


class Galaxy(Network):
    def __init__(self, user: Account, result_queue: Queue):
        super().__init__(user)
        self.user = user
        self.result_queue = result_queue
        logger.info("Galaxy instance created.")

    def prepare_fleet(self, task: Task) -> NetworkResponse:
        """
        Prepare fleet for the task.

        Args:
            task (Task): Task object.

        Returns:
            NetworkResponse: Contains arguments and token if successful.
        """
        PREPARE_FLEET_END_POINT = "game.php?page=my_fleet1"
        args = {}
        mission_mapping = {
            TaskType.ATTACK: MissionType.ATTACK,
            TaskType.EXPLORE: MissionType.EXPLORE
        }

        mission = mission_mapping.get(task.task_type)
        if not mission:
            err_msg = f"Unsupported task type: {task.task_type}"
            logger.error(err_msg)
            return NetworkResponse(status=-1, data={}, err_msg=err_msg)

        args.update({
            'mission': mission.value,
            'type': task.task_type.value,
            'galaxy': task.target.galaxy,
            'system': task.target.system,
            'planet': task.target.planet,
            'speed': 10
        })

        fleet_data = task.fleet.to_fleet()
        args.update(fleet_data)

        logger.info(f"Preparing fleet for task: {task}")
        response = self._post(PREPARE_FLEET_END_POINT, args)

        if response.status == 0:
            token = response.data.get('result', {}).get('token')
            if token:
                logger.info("Fleet prepared successfully.")
                return NetworkResponse(status=0, data={'args': args, 'token': token})
            else:
                err_msg = "Token not found in response."
                logger.error(err_msg)
                return NetworkResponse(status=-1, data={}, err_msg=err_msg)
        else:
            logger.error(f"Failed to prepare fleet: {response.err_msg}")
            return response

    def handle_attack_task(self, task: Task) -> NetworkResponse:
        """
        Handle an attack task.

        Args:
            task (Task): Task object.

        Returns:
            NetworkResponse: Contains the latest backtime if successful.
        """
        if not self.ready:
            err_msg = "Network not ready."
            logger.error(err_msg)
            return NetworkResponse(status=-1, data={}, err_msg=err_msg)

        total_finish_ts = -1
        for attempt in range(task.repeat):
            logger.info(f"Handling attack task attempt {attempt + 1}/{task.repeat}")
            response = self.handle_single_attack_task(task)
            if response.status != 0:
                logger.error("Error sending fleet.")
                # TODO: Update stats if necessary
            else:
                backtime = response.data.get('back_ts', -1)
                if backtime > total_finish_ts:
                    total_finish_ts = backtime
            time.sleep(1)
        return NetworkResponse(status=0, data={'total_finish_ts': total_finish_ts})

    def handle_single_attack_task(self, task: Task) -> NetworkResponse:
        """
        Handle a single attack task.

        Args:
            task (Task): Task object.

        Returns:
            NetworkResponse: Contains backtime if successful, error otherwise.
        """
        response = self.prepare_fleet(task)
        if response.status != 0:
            logger.error("Error preparing fleet.")
            return response

        args = response.data['args']
        token = response.data['token']
        SEND_FLEET_END_POINT = "game.php?page=fleet3"
        args['token'] = token

        logger.info("Sending fleet.")
        send_response = self._post(SEND_FLEET_END_POINT, args)

        if send_response.status != 0:
            logger.error(f"Error sending fleet: {send_response.err_msg}")
            return send_response

        logger.info("Fleet sent successfully.")
        backtime = send_response.data.get('result', {}).get('back_ts', -1)
        return NetworkResponse(status=0, data={'back_ts': backtime})

    def handle_explore_task(self, task: Task) -> NetworkResponse:
        """
        Handle an explore task.

        Args:
            task (Task): Task object.

        Returns:
            NetworkResponse: Contains the latest backtime if successful.
        """
        if not self.ready:
            err_msg = "Network not ready."
            logger.error(err_msg)
            return NetworkResponse(status=-1, data={}, err_msg=err_msg)

        total_finish_ts = -1
        for attempt in range(task.repeat):
            logger.info(f"Handling explore task attempt {attempt + 1}/{task.repeat}")
            response = self.prepare_fleet(task)
            if response.status != 0:
                logger.error("Error preparing fleet.")
                return response

            args = response.data['args']
            token = response.data['token']
            END_POINT = "game.php?page=fleet3"
            args['token'] = token
            args['staytime'] = 1

            logger.info("Sending exploration fleet.")
            send_response = self._post(END_POINT, args)

            if send_response.status != 0:
                logger.error(f"Error sending exploration fleet: {send_response.err_msg}")
                continue

            backtime = send_response.data.get('result', {}).get('back_ts', -1)
            if backtime > total_finish_ts:
                total_finish_ts = backtime

            logger.info("Exploration fleet sent successfully.")
            time.sleep(1)
        return NetworkResponse(status=0, data={'total_finish_ts': total_finish_ts})

    def handle_escape_task(self, task: Task) -> NetworkResponse:
        """
        Handle an escape task.

        Args:
            task (Task): Task object.

        Returns:
            NetworkResponse: Contains the result of the escape task.
        """
        logger.info("Handling escape task...")
        # TODO: Implement escape task handling
        return NetworkResponse(status=0, data={'message': 'Escape task not implemented yet.'})
