// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/account/user/{id}": {
            "get": {
                "description": "Get Account by User ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "account"
                ],
                "summary": "Get account by User ID",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "User ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successful response with account data",
                        "schema": {
                            "$ref": "#/definitions/account.userAccountResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request with error message",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found with error message",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error with error message",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/account/{id}": {
            "get": {
                "description": "Get Account by ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "account"
                ],
                "summary": "Get account by ID",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Account ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successful response with account data",
                        "schema": {
                            "$ref": "#/definitions/account.accountResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request with error message",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found with error message",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error with error message",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/ping": {
            "get": {
                "description": "do ping",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "example"
                ],
                "summary": "ping example",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api.responseMessage"
                        }
                    }
                }
            }
        },
        "/user": {
            "put": {
                "description": "Update User",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Update user",
                "parameters": [
                    {
                        "description": "User ID or Username",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.User"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successful response with user data",
                        "schema": {
                            "$ref": "#/definitions/user.userResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request with error message",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error with error message",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    }
                }
            },
            "post": {
                "description": "Create User",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Crea user",
                "parameters": [
                    {
                        "description": "User ID or Username",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.User"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successful response with user data",
                        "schema": {
                            "$ref": "#/definitions/user.userResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request with error message",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error with error message",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    }
                }
            },
            "delete": {
                "description": "Delete User",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Delete user",
                "parameters": [
                    {
                        "description": "User ID or Username",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.User"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successful response with user data",
                        "schema": {
                            "$ref": "#/definitions/user.userResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request with error message",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error with error message",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/user/balance": {
            "put": {
                "description": "Update a user balance",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Update a user balance",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "User ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successful response with user data",
                        "schema": {
                            "$ref": "#/definitions/user.userResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request with error message",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found with error message",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error with error message",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/user/{id}": {
            "get": {
                "description": "Get User by ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Get user by ID",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "User ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successful response with user data",
                        "schema": {
                            "$ref": "#/definitions/user.userResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request with error message",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error with error message",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/users": {
            "get": {
                "description": "Get all Users",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Get all users",
                "responses": {
                    "200": {
                        "description": "Successful response with user data",
                        "schema": {
                            "$ref": "#/definitions/user.usersResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request with error message",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error with error message",
                        "schema": {
                            "$ref": "#/definitions/api.ErrorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "account.accountResponse": {
            "type": "object",
            "properties": {
                "data": {
                    "$ref": "#/definitions/models.AccountDTO"
                },
                "succeed": {
                    "type": "boolean"
                },
                "traceID": {
                    "type": "string"
                }
            }
        },
        "account.userAccountResponse": {
            "type": "object",
            "properties": {
                "data": {
                    "$ref": "#/definitions/models.UserDTO"
                },
                "succeed": {
                    "type": "boolean"
                },
                "traceID": {
                    "type": "string"
                }
            }
        },
        "api.ErrorResponse": {
            "type": "object",
            "properties": {
                "error": {
                    "type": "string"
                },
                "message": {
                    "type": "string"
                },
                "succeed": {
                    "type": "boolean"
                },
                "traceID": {
                    "type": "string"
                }
            }
        },
        "api.responseMessage": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                }
            }
        },
        "gorm.DeletedAt": {
            "type": "object",
            "properties": {
                "time": {
                    "type": "string"
                },
                "valid": {
                    "description": "Valid is true if Time is not NULL",
                    "type": "boolean"
                }
            }
        },
        "models.Account": {
            "type": "object",
            "properties": {
                "createdAt": {
                    "type": "string"
                },
                "deletedAt": {
                    "$ref": "#/definitions/gorm.DeletedAt"
                },
                "email": {
                    "type": "string"
                },
                "expireAt": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "password": {
                    "description": "MD5 hash",
                    "type": "string"
                },
                "routeTasks": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.RouteTask"
                    }
                },
                "server": {
                    "type": "string"
                },
                "updatedAt": {
                    "type": "string"
                },
                "userID": {
                    "type": "integer"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "models.AccountDTO": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string"
                },
                "expireAt": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "route_tasks": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.RouteTask"
                    }
                },
                "server": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "models.Fleet": {
            "type": "object",
            "properties": {
                "createdAt": {
                    "type": "string"
                },
                "deletedAt": {
                    "$ref": "#/definitions/gorm.DeletedAt"
                },
                "id": {
                    "type": "integer"
                },
                "routeTasks": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.RouteTask"
                    }
                },
                "updatedAt": {
                    "type": "string"
                }
            }
        },
        "models.RouteTask": {
            "type": "object",
            "properties": {
                "accountID": {
                    "type": "integer"
                },
                "createdAt": {
                    "type": "string"
                },
                "deletedAt": {
                    "$ref": "#/definitions/gorm.DeletedAt"
                },
                "enabled": {
                    "type": "boolean"
                },
                "fleets": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.Fleet"
                    }
                },
                "from": {
                    "$ref": "#/definitions/models.Star"
                },
                "id": {
                    "type": "integer"
                },
                "logs": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.taskLog"
                    }
                },
                "name": {
                    "type": "string"
                },
                "next_start": {
                    "type": "string"
                },
                "repeat": {
                    "type": "integer"
                },
                "to": {
                    "$ref": "#/definitions/models.Star"
                },
                "updatedAt": {
                    "type": "string"
                }
            }
        },
        "models.Star": {
            "type": "object",
            "properties": {
                "galaxy": {
                    "description": "gorm.Model // NOTE: Is this necessary?",
                    "type": "integer"
                },
                "is_moon": {
                    "type": "boolean"
                },
                "location": {
                    "type": "integer"
                },
                "solar": {
                    "type": "integer"
                }
            }
        },
        "models.User": {
            "type": "object",
            "properties": {
                "accounts": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.Account"
                    }
                },
                "balance": {
                    "type": "integer"
                },
                "createdAt": {
                    "type": "string"
                },
                "deletedAt": {
                    "$ref": "#/definitions/gorm.DeletedAt"
                },
                "id": {
                    "type": "integer"
                },
                "password": {
                    "description": "WARNING: USERNAME MAY BE NOT UNIQUE! RECHECK THIS!\nNOTE: Checked in db, DO api check",
                    "type": "string"
                },
                "updatedAt": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "models.UserDTO": {
            "type": "object",
            "properties": {
                "accounts": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.AccountDTO"
                    }
                },
                "balance": {
                    "type": "integer"
                },
                "id": {
                    "type": "integer"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "models.taskLog": {
            "type": "object",
            "properties": {
                "createdAt": {
                    "type": "string"
                },
                "deletedAt": {
                    "$ref": "#/definitions/gorm.DeletedAt"
                },
                "id": {
                    "type": "integer"
                },
                "referID": {
                    "description": "引用的 Task ID",
                    "type": "integer"
                },
                "referType": {
                    "description": "引用的 Task 类型",
                    "type": "string"
                },
                "updatedAt": {
                    "type": "string"
                }
            }
        },
        "user.userResponse": {
            "type": "object",
            "properties": {
                "data": {
                    "$ref": "#/definitions/models.UserDTO"
                },
                "succeed": {
                    "type": "boolean"
                }
            }
        },
        "user.usersResponse": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.UserDTO"
                    }
                },
                "succeed": {
                    "type": "boolean"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}