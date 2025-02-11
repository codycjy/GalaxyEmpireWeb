import pytest
import json
import allure
import requests
from datetime import datetime

# Fixtures
@pytest.fixture
def base_url():
    """基础URL"""
    return "http://localhost:9333/api/v1/"

@pytest.fixture
def register_user(base_url):
    """注册测试用户（如果不存在）"""
    register_data1 = {
        "username": "task1",
        "password": "12345678"
    }
    register_data2 = {
        "username": "task2",
        "password": "12345678"
    }
    response1 = requests.post(
        f"{base_url}register",
        json=register_data1,
    )
    if response1.status_code == 200:
        allure.attach(response1.text, name="Register Task User 1", attachment_type=allure.attachment_type.TEXT)
    else:
        allure.attach(response1.text, name="Task User 1 Already Registered", attachment_type=allure.attachment_type.TEXT)

    response2 = requests.post(
        f"{base_url}register",
        json=register_data2,
    )
    if response2.status_code == 200:
        allure.attach(response2.text, name="Register Task User 2", attachment_type=allure.attachment_type.TEXT)
    else:
        allure.attach(response2.text, name="Task User 2 Already Registered", attachment_type=allure.attachment_type.TEXT)

@pytest.fixture
def login_user(base_url,register_user):
    """登录主用户并返回 session 和 token"""
    login_data = {
        "username": "task1",
        "password": "12345678"
    }
    response = requests.post(
        f"{base_url}login",
        json=login_data,
    )
    assert response.status_code == 200, "登录失败"

    token = response.json()['token']
    user_id = response.json()['user']['id']

    session = requests.Session()
    session.headers.update({
        "Authorization": token,
        "Content-Type": "application/json"
    })

    return {
        'session': session,
        'token': token,
        'user_id': user_id
    }

@pytest.fixture
def login_another_user(base_url):
    """登录另一个用户"""
    login_data = {
        "username": "task2",
        "password": "12345678"
    }
    response = requests.post(
        f"{base_url}login",
        json=login_data,
    )
    assert response.status_code == 200, "登录失败"

    token = response.json()['token']
    user_id = response.json()['user']['id']

    session = requests.Session()
    session.headers.update({
        "Authorization": token,
        "Content-Type": "application/json"
    })

    return {
        'session': session,
        'token': token,
        'user_id': user_id
    }

@pytest.fixture
def create_account(base_url, login_user):
    """为测试用户创建一个账户"""
    account_data = {
        "username": "testaccount",
        "password": "testpassword",
        "server": "ze",
        "email": "test@example.com"
    }
    response = login_user['session'].post(
        f"{base_url}account",
        json=account_data,
        headers={"Authorization": login_user['token']}
    )
    if response.status_code != 200:
        # Account already exists
        # find the account by username
        response = login_user['session'].get(
                base_url + "account/user/" + str(login_user['user_id']),)
        response.raise_for_status()
        data = response.json()
        for account in data['data']['accounts']:
            if account['username'] == account_data['username']:
                return account

    return response.json()['data']

@pytest.fixture
def create_test_task(base_url, login_user, create_account):
    """创建用于测试的临时任务"""
    account_id = create_account['id']
    task_data = {
        "name": f"测试任务_{datetime.now().strftime('%H%M%S')}",
        "task_type": 1,
        "start_planet": {"galaxy": 1, "system": 2, "planet": 3},
        "account_id": account_id
    }
    response = login_user['session'].post(
        f"{base_url}task",
        json=task_data,
        headers={"Authorization": login_user['token']}
    )
    assert response.status_code == 200, "创建任务失败"
    return response.json()['data']

# Test Cases
@allure.feature("任务模块")
class TestTask:
    @allure.story("创建任务")
    @pytest.mark.parametrize("caseType, task_data, expected_status", [
        ("Valid Task", {
            "name": "基础任务",
            "task_type": 1,
            "start_planet": {"galaxy": 1, "system": 2, "planet": 3, "is_moon": False},
            "account_id": 1  # 假设 account_id 为 1
        }, 200),
        ("Invalid Task Type", {"name": "无效类型", "task_type": "asddsda", "account_id": 1}, 400),
    ])
    def test_create_task(self, base_url, login_user, create_account, caseType, task_data, expected_status):
        """测试创建任务功能"""
        allure.dynamic.title(caseType)
        with allure.step("生成动态任务名称"):
            if "name" in task_data:
                task_data["name"] = f"{task_data['name']}_{datetime.now().strftime('%H%M%S')}"

        with allure.step("发送创建请求"):
            task_data["account_id"] = create_account['id']  # 使用创建的账户 ID
            response = login_user['session'].post(
                f"{base_url}task",
                json=task_data,
                headers={"Authorization": login_user['token']}
            )
            allure.attach(response.text, name="Response", attachment_type=allure.attachment_type.TEXT)

        with allure.step("验证响应"):
            assert response.status_code == expected_status, f"预期{expected_status}，实际{response.status_code}"

    @allure.story("更新任务")
    @pytest.mark.parametrize("caseType, update_data", [
        ("Update Start Planet", {"start_planet": {"galaxy": 2, "system": 3, "planet": 4, "is_moon": True}}),
        ("Update Fleet", {"fleet": {"lf": 10, "hf": 5, "ds": 30}}),
        ("Update Target", {"targets": [{"galaxy": 999, "system": 22, "planet": 8}]}),
        ("Update Target2", {"targets": [{"galaxy": 999, "system": 22, "planet": 8},{"galaxy": 999, "system": 22, "planet": 9}]}),
        ("Update inline value", { "name": "综合更新", "repeat": 3, })
    ])
    def test_update_task(self, base_url, login_user, create_account,create_test_task, caseType, update_data):
        """测试更新任务不同字段"""
        print(create_test_task)
        task_id = create_test_task['ID']
        original_name = create_test_task['name']

        with allure.step("执行更新操作"):
            # 生成动态数据
            if "name" in update_data:
                update_data["name"] = f"{update_data['name']}_{datetime.now().strftime('%H%M%S')}"

            response = login_user['session'].put(
                f"{base_url}task/{task_id}",
                json=update_data,
                headers={"Authorization": login_user['token']}
            )
            allure.attach(response.text, name="Update Response", attachment_type=allure.attachment_type.TEXT)
            assert response.status_code == 200, f"更新失败: {response.text}"

        with allure.step("验证更新结果"):
            print(json.dumps(create_account, indent=4, ensure_ascii=False))
            get_response = login_user['session'].get(
                f"{base_url}account/{create_account['id']}",
                headers={"Authorization": login_user['token']}
            )
            allure.attach(get_response.text, name="Get Response", attachment_type=allure.attachment_type.TEXT)

            tasks = get_response.json()['data']['tasks']
            print(task_id,"!!!!!!!!")
            print(json.dumps(tasks, indent=4, ensure_ascii=False))
            updated_task = next((t for t in tasks if t['ID'] == task_id), None)
            assert updated_task is not None, "更新后的任务未找到"

            # 字段验证逻辑
            for key in update_data:
                if key == 'start_planet':
                    for sub_key in update_data[key]:
                        assert updated_task['start_planet'][sub_key] == update_data[key][sub_key], f"{sub_key} 更新不匹配"
                elif key == 'fleet':
                    for ship in update_data[key]:
                        assert updated_task['fleet'][ship] == update_data[key][ship], f"{ship} 数量不匹配"
                elif key == 'targets':
                    assert len(updated_task['targets']) == len(update_data[key]), "目标数量不一致"
                else:
                    assert str(updated_task[key]) == str(update_data[key]), f"{key} 字段未正确更新"

    @allure.story("跨用户访问验证")
    def test_cross_user_access(self, base_url, login_user, login_another_user,create_account,create_test_task):
        """验证其他用户无法访问/修改本用户任务"""
        with allure.step("主用户创建任务"):
            task_data = {
                "name": f"权限验证_{datetime.now().strftime('%H%M%S')}",
                "task_type": 1,
                "start_planet": {"galaxy": 1, "system": 1, "planet": 1},
                "account_id": create_test_task['account_id']
            }
            create_res = login_user['session'].post(
                f"{base_url}task",
                json=task_data,
                headers={"Authorization": login_user['token']}
            )
            print(create_res.json())
            task_id = create_res.json()['data']['ID']

        with allure.step("其他用户尝试访问"):
            access_res = login_another_user['session'].get(
                f"{base_url}task/{task_id}",
                headers={"Authorization": login_another_user['token']}
            )
            assert access_res.status_code in [403, 404], "权限验证未生效"

        with allure.step("其他用户尝试修改"):
            update_res = login_another_user['session'].put(
                f"{base_url}task/{task_id}",
                json={"name": "非法修改"},
                headers={"Authorization": login_another_user['token']}
            )
            assert update_res.status_code in [403, 404], "非法修改未阻止"

    @allure.story("删除任务")
    def test_delete_task(self, base_url, login_user,create_account, create_test_task):
        """测试删除任务功能"""
        task_id = create_test_task['ID']

        with allure.step("删除任务"):
            delete_response = login_user['session'].delete(
                f"{base_url}task",
                json={"id": task_id},
            )
            allure.attach(delete_response.text, name="Delete Response", attachment_type=allure.attachment_type.TEXT)
            assert delete_response.status_code == 200, "删除失败"

        with allure.step("验证任务已删除"):
            get_response = login_user['session'].get(
                f"{base_url}account/{create_account['id']}",
                headers={"Authorization": login_user['token']}
            )
            tasks = get_response.json()['data']['tasks']
            assert task_id not in [t['ID'] for t in tasks], "任务未成功删除"
