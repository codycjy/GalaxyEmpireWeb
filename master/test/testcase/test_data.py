# testcase/test_data.py

auth_test_data = {
    'test_login': [
        ("Valid login", "testuser_valid", "12345678", "Login successful."),
        ("Invalid password", "testuser_invalid_pw", "wrongpassword", "Invalid credentials."),
        ("Nonexistent user", "nonexistent_user", "12345678", "User does not exist."),
    ],
    'test_register': [
        ("Valid registration", "newuser_valid", "password123", "newuser_valid@example.com", "Registration successful."),
        ("Existing username", "testuser1", "password123", "testuser1@example.com", "Username already exists."),
        ("Missing password", "newuser_no_pw", "", "", "Password is required."),
    ],
    'test_token': [
        ("Valid token", "1"),
        ("No token", "3"),
        ("Invalid token", "4"),
        ("Expired token", "2"),
    ],
    'test_create_account': [
        ("Valid Account",
         {"email": "test3@example.com",
          "password": "password123",
          "username": "testaccount3",
          },
         200),
        ("Invalid Email",
         {"email": "invalid",
          "password": "password123",
          "username": "testaccount4",
          },
         400),
        ("Missing Password",
         {"email": "test5@example.com"},
         400),
    ]
}
