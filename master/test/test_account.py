# test_account.py

import allure
import pytest

from testcase.test_data import auth_test_data

@allure.feature("账户模块")
class TestAccount:
    @allure.story("根据用户id查询账户")
    @pytest.mark.parametrize("caseType, user_key", [
        ("User Exists", "main"),
        ("User Does Not Exist", "nonexistent"),
        ("Bad Request", "invalid")
    ])
    def test_get_account_by_user_id(self, base_url, register_user, caseType, user_key):
        """
        根据用户ID查询账户信息。
        """
        allure.dynamic.title(caseType)
        allure.dynamic.description(f"测试用例类型: {caseType}")

        if user_key == "main":
            user_id = register_user['id']
            endpoint = f"{base_url}account/user/{user_id}"

            # **重要**：在查询之前，为用户创建一个account
            # 这里假设创建account的API为POST /account
            account_data = {
                "username": "testaccount_main",
                "email": "testaccount_main@example.com",
                "password": "password123",
            }
            create_response = register_user['session'].post(f"{base_url}account", json=account_data)
            allure.attach(create_response.text, name="Create Account Response", attachment_type=allure.attachment_type.TEXT)
            assert create_response.status_code == 200, f"主用户创建账户失败: {create_response.text}"
            account_id = create_response.json().get('data').get('id')
            assert account_id is not None, "创建账户响应中未包含账户ID"

        elif user_key == "nonexistent":
            user_id = "999"
            endpoint = f"{base_url}account/user/{user_id}"
        else:
            endpoint = f"{base_url}account/user/invalid"

        response = register_user['session'].get(endpoint)
        allure.attach(response.text, name="Response Data", attachment_type=allure.attachment_type.TEXT)

        if caseType == "User Exists":
            assert response.status_code == 200, "存在的用户ID应返回200"
            # 可进一步验证返回的account数据
            res_json = response.json()
            assert isinstance(res_json, list) or isinstance(res_json, dict), "返回的数据格式应为列表或字典"
            # 根据实际API返回格式进行断言，例如检查返回的账户是否包含刚创建的账户
            # 这里假设返回的是账户列表
            if isinstance(res_json, list):
                assert len(res_json) > 0, "用户应至少有一个账户"
            elif isinstance(res_json, dict):
                for account in res_json['data']['accounts']:
                    assert 'id' in account and account['id'] is not None, "账户应包含有效的ID"
        elif caseType == "User Does Not Exist":
            assert response.status_code == 404, "不存在的用户ID应返回404"
        else:
            assert response.status_code == 400, "无效的用户ID应返回400"

    @allure.story("创建账户")
    @pytest.mark.parametrize("caseType, account_data, expected_status", auth_test_data['test_create_account'])
    def test_create_account(self, base_url, register_user, caseType, account_data, expected_status):
        """
        测试创建账户功能。
        """
        allure.dynamic.title(caseType)
        allure.dynamic.description(f"测试用例类型: {caseType}")

        response = register_user['session'].post(f"{base_url}account", json=account_data)
        allure.attach(response.text, name="Response Data", attachment_type=allure.attachment_type.TEXT)

        assert response.status_code == expected_status, f"预期状态码: {expected_status}, 实际状态码: {response.status_code}"
        if expected_status == 200:
            res_json = response.json()
            assert res_json.get('data').get('id') is not None, "创建账户应返回账户ID"
        else:
            res_json = response.json()
            assert 'message' in res_json, "错误响应应包含消息"

    @allure.story("确保交叉用户无法访问其他用户的账户")
    def test_cross_user_access(self, base_url, register_user, register_cross_user):
        """
        确保交叉用户无法访问其他用户的账户信息。
        """
        main_user = register_user
        cross_user = register_cross_user

        # 主用户创建一个账户
        account_data = {
            "username": "main_account",
            "email": "main_account@example.com",
            "password": "password123",
        }
        create_response = main_user['session'].post(f"{base_url}account", json=account_data)
        allure.attach(create_response.text, name="Create Account Response", attachment_type=allure.attachment_type.TEXT)
        assert create_response.status_code == 200, f"主用户创建账户失败: {create_response.text}"
        account_id = create_response.json().get('data').get('id')
        assert account_id is not None, "创建账户响应中未包含账户ID"

        # 交叉用户尝试访问主用户的账户
        access_response = cross_user['session'].get(f"{base_url}account/{account_id}")
        allure.attach(access_response.text, name="Access Response Data", attachment_type=allure.attachment_type.TEXT)
        assert access_response.status_code == 403, "交叉用户应无权限访问主用户的账户"
