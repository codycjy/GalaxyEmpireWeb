import time

import allure
import pytest
import requests
from testcase.conftest import auth_test_data


@allure.feature("登录注册模块")
class Test_Login():
    def setup_class(self):
        self.root_url = "http://localhost:9333/api/v1/"
        self.req_session = requests.Session()
        # 添加测试用户
        rep0 = self.req_session.post(self.root_url + 'register', json={"username": 'testuser1', "password": '123456'})
        # 获取token
        rep = self.req_session.post(self.root_url + 'login', json={"username": 'testuser1', "password": '123456'})
        self.token = rep.json()['token']

    @allure.story("登录")
    @pytest.mark.parametrize("caseType,username,password,msg", auth_test_data['test_login'])
    def test_login(self, caseType, username, password, msg):
        allure.dynamic.title(caseType)
        allure.dynamic.description(caseType)
        responce = self.req_session.post(self.root_url + 'login', json={"username": username, "password": password})
        allure.attach(responce.text, name="Response Data", attachment_type=allure.attachment_type.TEXT)
        res_json = responce.json()
        if responce.status_code == 200:
            assert res_json['token'] is not None
        else:
            assert res_json['message'] == msg

    @allure.story("注册")
    @pytest.mark.parametrize("caseType,username,password,msg", auth_test_data['test_register'])
    def test_register(self, caseType, username, password, msg):
        allure.dynamic.title(caseType)
        allure.dynamic.description(caseType)
        responce = self.req_session.post(self.root_url + 'register', json={"username": username, "password": password})
        allure.attach(responce.text, name="Response Data", attachment_type=allure.attachment_type.TEXT)
        if responce.status_code == 200:
            assert responce.json()['succeed'] is True
        else:
            assert responce.json()['message'] == msg

    @allure.story("token验证")
    @pytest.mark.parametrize("caseType,type", auth_test_data['test_token'])
    def test_token(self, caseType, type):
        allure.dynamic.title(caseType)
        allure.dynamic.description(caseType)
        token = self.token
        if type == '3':
            token = ''     # 无token
        elif type == '4':
            token = token + 'a'  # 错误token
        self.req_session.headers.update({
            'Authorization': f'{token}'
        })
        if type == '2':
            time.sleep(15)  # token过期
        responce = self.req_session.get(self.root_url + 'user/1')
        if type == '1':
            assert responce.status_code == 200
        else:
            assert responce.status_code == 401
        allure.attach(responce.text, name="Response Data", attachment_type=allure.attachment_type.TEXT)
