import allure
import pytest
import requests


@allure.feature("账户模块")
class Test_Account():
    @classmethod
    def setup_class(cls):  # Changed from self to cls for class method
        cls.base_url = 'http://localhost:9333/api/v1'
        cls.req_session = requests.session()
        # register account test user
        account_data = {
            "username": "testuser1_account",
            "password": "12345678"
        }

        cls.req_session.post(f'{cls.base_url}/register', json=account_data)

        # Login first
        login_data = {
            "username": "testuser1_account",
            "password": "12345678"
        }
        login_response = requests.post(f'{cls.base_url}/login', json=login_data)
        token = login_response.json().get('token')
        print(f"Token: {token}")
        cls.req_session.headers.update({
            'Authorization': f'{token}'
        })
        #添加测试账户
        test_account = self.req_session.post(self.root_url + '/account', json={"username": 'testaccount1', "password": '12345678'})

        # Create test accounts if needed
        cls.create_test_accounts()

    @classmethod  # Added classmethod decorator
    def create_test_accounts(cls):  # Changed from self to cls
        # Create test accounts
        account_data = {
            "username": "testaccount1",
            "email": "test1@example.com",
            "password": "password123",
        }
        cls.req_session.post(f'{cls.base_url}/account', json=account_data)

        account_data2 = {
            "username": "testaccount2",
            "email": "test2@example.com",
            "password": "password123",
        }
        cls.req_session.post(f'{cls.base_url}/account', json=account_data2)

    @allure.story("根据用户id查询账户")
    @pytest.mark.parametrize("caseType, userId", [
        ("user exist", "1"),
        ("user not exist", "999"),
        ("bad request", "invalid")
    ])
    def test_with_getAccountByUserId(self, caseType, userId):
        allure.dynamic.title(caseType)
        allure.dynamic.description(caseType)
        print(self.req_session.headers.get('Authorization'))
        response = self.req_session.get(f'{self.base_url}/account/user/{userId}')
        allure.attach(response.text, name="Response Data", attachment_type=allure.attachment_type.TEXT)
        if response.status_code == 200:
            assert response.json()['succeed'] is True
        else:
            assert response.json()['message'] == msg

        if caseType == "user exist":
            assert response.status_code == 200
        elif caseType == "user not exist":
            assert response.status_code == 404
        else:
            assert response.status_code == 400

    @allure.story("根据id查询账户")
    @pytest.mark.parametrize("caseType, accountId", [
        ("account exist", "1"),
        ("account not exist", "999"),
        ("bad request", "invalid")
    ])
    def test_getAccountById(self, caseType, accountId):
        allure.dynamic.title(caseType)
        allure.dynamic.description(caseType)
        response = self.req_session.get(f'{self.base_url}/account/{accountId}')
        allure.attach(response.text, name="Response Data", attachment_type=allure.attachment_type.TEXT)
        if response.status_code == 200:
            assert response.json()['succeed'] is True
        else:
            assert response.json()['message'] == msg

        if caseType == "account exist":
            assert response.status_code == 200
        elif caseType == "account not exist":
            assert response.status_code == 403
        else:
            assert response.status_code == 400

    @allure.story("创建账户")
    @pytest.mark.parametrize("caseType, account_data", [
        ("valid account",
         {"email": "test3@example.com",
          "password": "password123",
          "username": "testaccount3",
          }),
        ("invalid email", {"email": "invalid",
                           "password": "password123",
                           "username": "testaccount4", }),
        ("missing password", {"email": "test5@example.com"}),
    ])
    def test_createAccount(self, caseType, account_data):
        allure.dynamic.title(caseType)
        allure.dynamic.description(caseType)
        response = self.req_session.post(f'{self.base_url}/account', json=account_data)
        allure.attach(response.text, name="Response Data", attachment_type=allure.attachment_type.TEXT)

        if caseType == "valid account":
            assert response.status_code == 200
        else:
            assert response.status_code == 400
