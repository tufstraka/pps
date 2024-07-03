// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "API Support",
            "email": "keithkadima@gmail.com"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/login": {
            "post": {
                "description": "Authenticate a user and return a JWT token",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Login a user",
                "parameters": [
                    {
                        "description": "User details",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/main.UserLogin"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Login successful with user details and token",
                        "schema": {
                            "$ref": "#/definitions/main.LoginSuccessResponse"
                        }
                    },
                    "401": {
                        "description": "Invalid credentials",
                        "schema": {
                            "$ref": "#/definitions/main.FailResponse"
                        }
                    },
                    "500": {
                        "description": "Server error",
                        "schema": {
                            "$ref": "#/definitions/main.FailResponse"
                        }
                    }
                }
            }
        },
        "/payments/initiate": {
            "post": {
                "description": "Initiate a payment for a user",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "payments"
                ],
                "summary": "Initiate a payment",
                "parameters": [
                    {
                        "description": "Payment Request",
                        "name": "payment",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/main.PaymentRequest"
                        }
                    }
                ],
                "responses": {
                    "202": {
                        "description": "Success",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/payments/send-to-mobile": {
            "post": {
                "description": "Send money to a mobile number via the Payd API",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "payments"
                ],
                "summary": "Send money to a mobile number",
                "parameters": [
                    {
                        "description": "Mobile Payment Request",
                        "name": "mobilePayment",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/main.MobilePaymentRequest"
                        }
                    }
                ],
                "responses": {
                    "202": {
                        "description": "Accepted",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/payments/status/{id}": {
            "get": {
                "description": "Get the status of a payment by ID",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "payments"
                ],
                "summary": "Get payment status",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Payment ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Accepted",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/register": {
            "post": {
                "description": "Register a new user with the provided details",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Register a new user",
                "parameters": [
                    {
                        "description": "User Details",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/main.User"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Registration successful",
                        "schema": {
                            "$ref": "#/definitions/main.SuccessResponse"
                        }
                    },
                    "401": {
                        "description": "Registration failed",
                        "schema": {
                            "$ref": "#/definitions/main.FailResponse"
                        }
                    },
                    "500": {
                        "description": "Server error",
                        "schema": {
                            "$ref": "#/definitions/main.FailResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "main.FailResponse": {
            "type": "object",
            "properties": {
                "status": {
                    "type": "string"
                }
            }
        },
        "main.LoginSuccessResponse": {
            "type": "object",
            "properties": {
                "status": {
                    "type": "string"
                },
                "token": {
                    "type": "string"
                }
            }
        },
        "main.MobilePaymentRequest": {
            "type": "object",
            "properties": {
                "account_id": {
                    "type": "string"
                },
                "amount": {
                    "type": "number"
                },
                "callback_url": {
                    "type": "string"
                },
                "channel": {
                    "type": "string"
                },
                "narration": {
                    "type": "string"
                },
                "payment_method": {
                    "type": "string"
                },
                "phone_number": {
                    "type": "string"
                }
            }
        },
        "main.PaymentRequest": {
            "type": "object",
            "properties": {
                "amount": {
                    "type": "number"
                },
                "email": {
                    "type": "string"
                },
                "first_name": {
                    "type": "string"
                },
                "last_name": {
                    "type": "string"
                },
                "location": {
                    "type": "string"
                },
                "payment_method": {
                    "type": "string"
                },
                "phone": {
                    "type": "string"
                },
                "reason": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "main.SuccessResponse": {
            "type": "object",
            "properties": {
                "user": {
                    "$ref": "#/definitions/main.User"
                }
            }
        },
        "main.User": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string"
                },
                "location": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "phone": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "main.UserLogin": {
            "type": "object",
            "properties": {
                "password": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "0.1",
	Host:             "54.145.134.156:8083",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "Payment Gateway API",
	Description:      "This is a payment gateway service that integrates with the authentication and payments services.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
