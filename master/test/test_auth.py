# test_auth.py

import allure
import pytest
import time

from testcase.test_data import auth_test_data

@allure.feature("登录注册模块")
class TestAuth:
    @allure.story("登录")
    @pytest.mark.parametrize("caseType,username,password,msg", auth_test_data['test_login'])
    def test_login(self, base_url, session, caseType, username, password, msg):
        """
        测试用户登录功能。
        """
        allure.dynamic.title(caseType)
        allure.dynamic.description(f"测试用例类型: {caseType}")

        # 登录请求
        response = session.post(f"{base_url}login", json={"username": username, "password": password})
        allure.attach(response.text, name="Response Data", attachment_type=allure.attachment_type.TEXT)
        res_json = response.json()

        if response.status_code == 200:
            assert res_json.get('token') is not None, "Token 应存在于响应中"
        else:
            assert res_json.get('token') is None, f"Token 不应存在于响应中: {res_json.get('message')}"

    @allure.story("注册")
    @pytest.mark.parametrize("caseType,username,password,email,msg", auth_test_data['test_register'])
    def test_register(self, base_url, session, caseType, username, password, email, msg):
        """
        测试用户注册功能。
        """
        allure.dynamic.title(caseType)
        allure.dynamic.description(f"测试用例类型: {caseType}")

        registration_data = {"username": username, "password": password}
        if email:
            registration_data["email"] = email

        response = session.post(f"{base_url}register", json=registration_data)
        allure.attach(response.text, name="Response Data", attachment_type=allure.attachment_type.TEXT)
        res_json = response.json()

        if response.status_code == 200:
            assert res_json.get('succeed') is True, "注册应成功"
        else:
            assert res_json.get('succeed') is False, f"注册不应成功: {res_json.get('message')}"

    @allure.story("Token验证")
    @pytest.mark.parametrize("caseType,token_type", auth_test_data['test_token'])
    def test_token_validation(self, base_url, session, register_user, caseType, token_type):
        """
        测试Token验证功能，包括有效Token、无Token、无效Token和过期Token。
        """
        allure.dynamic.title(caseType)
        allure.dynamic.description(f"测试用例类型: {caseType}")

        headers = {}
        if token_type == '1':
            # 有效Token
            headers['Authorization'] = register_user['token']
        elif token_type == '2':
            # 模拟Token过期，可以通过等待或其他方式模拟
            headers['Authorization'] = register_user['token']
            # 这里假设通过某种方式使Token过期，如等待一定时间
            # 例如，等待10秒
            time.sleep(10)  # 假设Token在10秒后过期
        elif token_type == '3':
            # 无Token
            session.headers.pop('Authorization', None)
        elif token_type == '4':
            # 无效Token
            session.headers['Authorization'] = register_user['token'] + 'a'

        response = session.get(f"{base_url}user/{register_user['id']}", headers=headers)
        allure.attach(response.text, name="Response Data", attachment_type=allure.attachment_type.TEXT)

        if token_type == '1':
            assert response.status_code == 200, "有效Token应返回200"
        elif token_type == '2':
            # 假设Token过期返回401
            assert response.status_code == 401, "过期Token应返回401"
        else:
            # 无Token或无效Token应返回401
            assert response.status_code == 401, "无效或缺失Token应返回401"
