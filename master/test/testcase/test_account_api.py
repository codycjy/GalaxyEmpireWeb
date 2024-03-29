import os
import allure
import pytest
import requests
from testcase.conftest import account_test_data

@allure.feature("账户模块")
class Test_Account():
    def setup_class(self):
        self.root_url = 'http://localhost:9333/api/v1'
        self.req_session = requests.session()
        #测试用户
        rep0 = self.req_session.post(self.root_url + '/register', json={"username": 'testuser3', "password": '12345678'})
        # 获取token
        rep = self.req_session.post(self.root_url + '/login', json={"username": 'testuser3', "password": '12345678'})

        self.token = rep.json()['token']
        self.req_session.headers.update({
            'Authorization': f'{self.token}'
        })
        #添加测试账户
        test_account = self.req_session.post(self.root_url + '/account', json={"username": 'testaccount1', "password": '12345678'})

    @allure.story("根据用户id查询账户")
    @pytest.mark.parametrize("caseType, userId, msg", account_test_data['test_getAccountsByUserId'])
    def test_with_getAccountByUserId(self, caseType, userId, msg):
        allure.dynamic.title(caseType)
        allure.dynamic.description(caseType)
        response = self.req_session.get(self.root_url + f'/account/user/{userId}')
        allure.attach(response.text, name="Response Data", attachment_type=allure.attachment_type.TEXT)
        if response.status_code == 200:
            assert response.json()['succeed'] is True
        else:
            assert response.json()['message'] == msg

    @allure.story("根据id查询账户")
    @pytest.mark.parametrize("caseType, accountId, msg", account_test_data['test_getAccountById'])
    def test_getAccountById(self, caseType, accountId, msg):
        allure.dynamic.title(caseType)
        allure.dynamic.description(caseType)
        response = self.req_session.get(self.root_url + "/account/" +  accountId)
        allure.attach(response.text, name="Response Data", attachment_type=allure.attachment_type.TEXT)
        if response.status_code == 200:
            assert response.json()['succeed'] is True
        else:
            assert response.json()['message'] == msg

