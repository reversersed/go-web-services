// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "consumes": [
        "application/json"
    ],
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
        "/users": {
            "get": {
                "description": "Get user using Id or login, both params are optional, but one of them is necessary",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Finds user by id or login",
                "parameters": [
                    {
                        "type": "string",
                        "description": "User id",
                        "name": "id",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "User login",
                        "name": "login",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successful response",
                        "schema": {
                            "$ref": "#/definitions/user.User"
                        }
                    },
                    "400": {
                        "description": "Returns when service didn't get a parameters",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "404": {
                        "description": "Returns when service can't find user by provided credentials (user not found)",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "500": {
                        "description": "Returns when there's some internal error that needs to be fixed",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    }
                }
            }
        },
        "/users/changename": {
            "patch": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "New login must be unique. Login changing are available only 1 time per month",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Update user's login",
                "parameters": [
                    {
                        "description": "New user login. Must be unique",
                        "name": "NewLogin",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/user.UpdateUserLoginQuery"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successful response. User's login was updated",
                        "schema": {
                            "$ref": "#/definitions/jwt.JwtResponse"
                        }
                    },
                    "401": {
                        "description": "Return's if service can't authorize user",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "403": {
                        "description": "Return's if user has login changing cooldown",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "404": {
                        "description": "Return's if user is not authorized",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "409": {
                        "description": "Return's if new user's login already taken",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "500": {
                        "description": "Returns when there's some internal error that needs to be fixed or smtp server is not responding",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "501": {
                        "description": "Returns if query was incorrect",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    }
                }
            }
        },
        "/users/delete": {
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Only user can delete his own account. To delete user he needs to confirm his password",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Deletes user's account",
                "parameters": [
                    {
                        "description": "User password",
                        "name": "Password",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/user.DeleteUserQuery"
                        }
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Successful response. User was deleted, need to remove his session"
                    },
                    "400": {
                        "description": "Return's if user typed incorrect password",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "401": {
                        "description": "Return's if service can't authorize user",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "404": {
                        "description": "Return's if user is not authorized",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "500": {
                        "description": "Returns when there's some internal error that needs to be fixed or smtp server is not responding",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "501": {
                        "description": "Returns if query was incorrect",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    }
                }
            }
        },
        "/users/email": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "If code field is empty: send or resend confirmation message to user's email\nMessage can be resended every 1 minutes\nIf code field is not empty: validate code and approve email, code is expired within 10 minutes",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Confirm user's email",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Confirmation code",
                        "name": "code",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successful response. Confirmation code was sent"
                    },
                    "204": {
                        "description": "Successful response. Email was confirmed"
                    },
                    "400": {
                        "description": "Return's if user's email already confirmed",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "401": {
                        "description": "Return's if service can't authorize user",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "403": {
                        "description": "Return's if email can't be resend now (cooldown still active)",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "404": {
                        "description": "Return's if user is authorized, but service can't identity him",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "500": {
                        "description": "Returns when there's some internal error that needs to be fixed or smtp server is not responding",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "501": {
                        "description": "Returns when provided confirmation code is incorrect or code is expired",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    }
                }
            }
        },
        "/users/login": {
            "post": {
                "description": "Finds user by login and password\nReturns a token and refresh token. Token expires in 1 hour, refresh token expires in 7 days and stores in cache (removing after system restart)\nLogin field can be provided with user login or email",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Authenticates user",
                "parameters": [
                    {
                        "description": "User credentials",
                        "name": "query",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/user.UserAuthQuery"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successful response. Returns user's login, roles and personal token and refresh token. Refresh token stores in cache",
                        "schema": {
                            "$ref": "#/definitions/jwt.JwtResponse"
                        }
                    },
                    "404": {
                        "description": "Returns when service can't find user by provided credentials (user not found)",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "500": {
                        "description": "Returns when there's some internal error that needs to be fixed",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "501": {
                        "description": "Returns when provided data was not validated",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    }
                }
            }
        },
        "/users/refresh": {
            "post": {
                "description": "Generate new token by provided refresh token\nRefresh token stored in cache and expires in 7 days. If system was restarted, all tokens are cleared and sessions are deleted",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Generate new token",
                "parameters": [
                    {
                        "description": "Request query with user's refresh token",
                        "name": "query",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/jwt.RefreshTokenQuery"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successful response. Returns the same data as in authorization",
                        "schema": {
                            "$ref": "#/definitions/jwt.JwtResponse"
                        }
                    },
                    "404": {
                        "description": "Returns when service can't find user by provided credentials (user not found)",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "500": {
                        "description": "Returns when there's some internal error that needs to be fixed",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "501": {
                        "description": "Returns when provided data was not validated",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    }
                }
            }
        },
        "/users/register": {
            "post": {
                "description": "Creates a new instance of user and returns authorization principals",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Register user",
                "parameters": [
                    {
                        "description": "User credentials",
                        "name": "query",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/user.UserRegisterQuery"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successful token response. Returns the same response as in authorization",
                        "schema": {
                            "$ref": "#/definitions/jwt.JwtResponse"
                        }
                    },
                    "409": {
                        "description": "Returns when there's already exist user with provided login",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "500": {
                        "description": "Returns when there's some internal error that needs to be fixed",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "501": {
                        "description": "Returns when provided data was not validated",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "errormiddleware.Error": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "string"
                },
                "dev_message": {
                    "type": "string"
                },
                "messages": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "jwt.JwtResponse": {
            "type": "object",
            "properties": {
                "login": {
                    "type": "string"
                },
                "refreshtoken": {
                    "type": "string"
                },
                "roles": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "token": {
                    "type": "string"
                }
            }
        },
        "jwt.RefreshTokenQuery": {
            "type": "object",
            "required": [
                "refreshtoken"
            ],
            "properties": {
                "refreshtoken": {
                    "type": "string"
                }
            }
        },
        "user.DeleteUserQuery": {
            "type": "object",
            "required": [
                "password"
            ],
            "properties": {
                "password": {
                    "type": "string"
                }
            }
        },
        "user.UpdateUserLoginQuery": {
            "type": "object",
            "required": [
                "newlogin"
            ],
            "properties": {
                "newlogin": {
                    "type": "string",
                    "maxLength": 16,
                    "minLength": 4
                }
            }
        },
        "user.User": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string"
                },
                "emailconfirmed": {
                    "type": "boolean"
                },
                "id": {
                    "type": "string"
                },
                "login": {
                    "type": "string"
                },
                "roles": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "user.UserAuthQuery": {
            "type": "object",
            "required": [
                "login",
                "password"
            ],
            "properties": {
                "login": {
                    "type": "string",
                    "example": "admin"
                },
                "password": {
                    "type": "string",
                    "example": "admin"
                }
            }
        },
        "user.UserRegisterQuery": {
            "type": "object",
            "required": [
                "email",
                "login",
                "password"
            ],
            "properties": {
                "email": {
                    "type": "string",
                    "example": "user@example.com"
                },
                "login": {
                    "type": "string",
                    "maxLength": 16,
                    "minLength": 4,
                    "example": "user"
                },
                "password": {
                    "type": "string",
                    "maxLength": 32,
                    "minLength": 8,
                    "example": "User!1password"
                }
            }
        }
    },
    "securityDefinitions": {
        "ApiKeyAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:9000",
	BasePath:         "/api/v1/",
	Schemes:          []string{},
	Title:            "API",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
