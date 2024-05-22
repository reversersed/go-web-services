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
        "/users/login": {
            "post": {
                "description": "Finds user by login and password\nReturns a token and refresh token. Token expires in 1 hour, refresh token expires in 7 days and stores in cache (removing after system restart)",
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
                            "$ref": "#/definitions/auth.UserAuthQuery"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/jwt.JwtResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "501": {
                        "description": "Not Implemented",
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
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/jwt.JwtResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    },
                    "501": {
                        "description": "Not Implemented",
                        "schema": {
                            "$ref": "#/definitions/errormiddleware.Error"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "auth.UserAuthQuery": {
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
        "errormiddleware.Error": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "string"
                },
                "dev_message": {
                    "type": "string"
                },
                "message": {
                    "type": "string"
                }
            }
        },
        "jwt.JwtResponse": {
            "type": "object",
            "properties": {
                "refreshtoken": {
                    "type": "string"
                },
                "token": {
                    "type": "string"
                },
                "username": {
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
