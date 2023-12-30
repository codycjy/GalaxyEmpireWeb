import os
import allure
import pytest
import requests
from testcase.conftest import account_test_data

@allure.feature("账户模块")
class Test_Account():
    def setup_class(self):
        token = 'nihao'
        self.root_url = 'http://localhost:9333/api/v1/account/'
        self.req_session = requests.session()
        self.req_session.headers.update({
            'Authorization': f'{token}'
        })

    @allure.story("根据用户id查询账户")
    @pytest.mark.parametrize("caseType, userId", account_test_data['test_getAccountsByUserId'])
    def test_with_getAccountByUserId(self, caseType, userId):
        allure.dynamic.title(caseType)
        allure.dynamic.description(caseType)
        response = self.req_session.get(self.root_url + f'user/{userId}')
        allure.attach(response.text, name="Response Data", attachment_type=allure.attachment_type.TEXT)

    @allure.story("根据id查询账户")
    @pytest.mark.parametrize("caseType, accountId", account_test_data['test_getAccountById'])
    def test_getAccountById(self, caseType, accountId):
        allure.dynamic.title(caseType)
        allure.dynamic.description(caseType)
        response = self.req_session.get(self.root_url + accountId)
        allure.attach(response.text, name="Response Data", attachment_type=allure.attachment_type.TEXT)

if __name__ == '__main__':
    pytest.main(["-s", "test_account_api.py", "--alluredir", '../result'])
    os.system('allure generate ../result -o ../report')
