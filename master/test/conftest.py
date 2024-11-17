# conftest.py

import pytest
import requests
import time
from testcase.test_data import auth_test_data

@pytest.fixture(scope="session")
def base_url():
    return "http://localhost:9333/api/v1/"

@pytest.fixture(scope="session")
def session():
    return requests.Session()

@pytest.fixture(scope="session")
def register_user(session, base_url):
    """
    注册一个用户，并返回用户信息，包括ID和Token。
    """
    prefix = "testuser_account"
    timestamp = str(int(time.time() * 1000))[-4:]
    username = f"{prefix}_{timestamp}"
    password = "12345678"
    email = f"{username}@example.com"

    registration_data = {
        "username": username,
        "password": password,
        "email": email
    }

    # 注册用户
    response = session.post(f"{base_url}register", json=registration_data)
    assert response.status_code == 200, f"用户注册失败: {response.text}"
    user_id = response.json().get('data', {}).get('id')
    assert user_id is not None, "注册响应中未包含用户ID"

    # 登录用户
    login_data = {
        "username": username,
        "password": password
    }
    login_response = session.post(f"{base_url}login", json=login_data)
    assert login_response.status_code == 200, f"登录失败: {login_response.text}"
    token = login_response.json().get('token')
    assert token is not None, "登录响应中未包含token"

    # 更新会话头，不使用 Bearer
    session.headers.update({'Authorization': f'{token}'})

    # 返回用户信息
    return {
        "id": user_id,
        "username": username,
        "token": token,
        "session": session
    }

@pytest.fixture(scope="session")
def register_cross_user(session, base_url):
    """
    注册另一个用户，用于测试交叉访问权限。
    """
    # 创建独立的会话
    cross_session = requests.Session()

    prefix = "testuser_account_cross"
    timestamp = str(int(time.time() * 1000))[-4:]
    username = f"{prefix}_{timestamp}"
    password = "87654321"
    email = f"{username}@example.com"

    registration_data = {
        "username": username,
        "password": password,
        "email": email
    }

    # 注册用户
    response = cross_session.post(f"{base_url}register", json=registration_data)
    assert response.status_code == 200, f"交叉用户注册失败: {response.text}"
    user_id = response.json().get('data', {}).get('id')
    assert user_id is not None, "交叉用户注册响应中未包含用户ID"

    # 登录用户
    login_data = {
        "username": username,
        "password": password
    }
    login_response = cross_session.post(f"{base_url}login", json=login_data)
    assert login_response.status_code == 200, f"交叉用户登录失败: {login_response.text}"
    token = login_response.json().get('token')
    assert token is not None, "交叉用户登录响应中未包含token"

    # 更新会话头，不使用 Bearer
    cross_session.headers.update({'Authorization': f'{token}'})

    # 返回用户信息
    return {
        "id": user_id,
        "username": username,
        "token": token,
        "session": cross_session
    }
