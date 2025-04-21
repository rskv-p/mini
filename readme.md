
# **Mini Framework** ![Go](https://cdn.jsdelivr.net/gh/golang/logo@master/256/golang-logo.png)

The **Mini Framework** is a modular application framework designed to simplify the development and scaling of your services. With a clean separation of concerns, it lets you build robust systems through **modules**—each containing **actions** that encapsulate the business logic. You can effortlessly manage services, handle APIs, interact with databases, and facilitate communication between services. 

---

## **Table of Contents** 📑
1. [Introduction](#introduction)
2. [Features](#features)
3. [Architecture](#architecture)
4. [Modules](#modules)
5. [Service Lifecycle](#service-lifecycle)
6. [How to Use](#how-to-use)
7. [API Documentation](#api-documentation)
8. [Running the Application](#running-the-application)
9. [Contributing](#contributing)
10. [License](#license)

---

## **Introduction** 

The **Mini Framework** provides an intuitive and modular approach to developing microservices. It integrates essential services like database management, messaging, logging, and authentication into self-contained modules. Each module houses actions (business logic), allowing you to dynamically load, initialize, and execute them to streamline your service development.

---

## **Features** 

- **Modular Design**: Encapsulates features like database management, API handling, and authentication into easily manageable modules.
- **Action-Based Logic**: Exposes actions as simple functions that execute business tasks, making them reusable and composable.
- **Service Lifecycle**: Seamlessly manage the startup and shutdown of services and modules, ensuring smooth operation.
- **Dynamic API Exposure**: Automatically exposes actions as APIs over HTTP, allowing quick integration with other services.
- **Database Integration**: Integrates with databases like SQLite and PostgreSQL via GORM ORM.
- **Inter-Service Communication**: Supports communication between services through a publish/subscribe bus.
- **Authentication & Authorization**: Built-in user authentication and role management.

---

## **Architecture** 

The **Mini Framework** follows a **three-tier architecture** for simplicity and scalability:

1. **Service**: The entry point to your application, responsible for initializing and managing modules.
2. **Module**: A self-contained unit that encapsulates related business logic and can contain multiple actions.
3. **Action**: The atomic units of work within a module that execute the actual business logic.

### **Service → Module → Action**

- **Service**: Starts and manages modules.
- **Module**: Defines the business logic for specific application functions (e.g., authentication, database).
- **Action**: Executes the specific tasks, such as creating a user or sending a message.

---

## **Modules** 

The **Mini Framework** comes pre-packaged with the following modules:

1. **`m_api`**: API management system for exposing actions over HTTP.
2. **`m_db`**: Database module for handling database operations using GORM.
3. **`m_bus`**: Messaging bus for inter-service communication using the publish/subscribe pattern.
4. **`m_log`**: Logging module for structured logging across services.
5. **`m_auth`**: User authentication and role management.
6. **`m_rtm`**: Runtime module for executing actions in various environments (e.g., Go, JavaScript).
7. **`m_cfg`**: Configuration management module to manage and publish configuration data.

Modules are extendable, and you can easily create new ones to fit your specific requirements.

---

## **Service Lifecycle** 

The lifecycle of the service consists of two main stages:

1. **Start**: 
   - Initializes and configures all modules.
   - Registers actions and sets up necessary resources (e.g., database connections, API server).
   
2. **Stop**: 
   - Gracefully shuts down all modules.
   - Cleans up resources (e.g., closes database connections).

---

## **How to Use** 

1. **Clone the repository**:
   ```bash
   git clone https://github.com/your-repo/mini.git
   cd mini
   ```

2. **Install dependencies**:
   Follow setup instructions for your environment (e.g., Go, GORM).

3. **Create a new module**:
   A module bundles related actions. Here’s an example:

   ```go
   package m_example

   import (
     "fmt"
     "github.com/rskv-p/mini/act"
     "github.com/rskv-p/mini/typ"
   )

   func ExampleModule() typ.IModule {
     return &mod.Module{
       ModName: "example",
       Acts: []typ.ActionDef{
         {
           Name: "example.action",
           Func: func(a typ.IAction) any {
             return "This is an example action"
           },
         },
       },
     }
   }
   ```

4. **Register the module** in `main.go`:

   ```go
   import "github.com/rskv-p/mini/mod/m_example"
   
   service := &srv.Service{
     Name: "ExampleService",
   }

   service.AddModule(m_example.ExampleModule())

   if err := service.Start(); err != nil {
     log.Fatalf("Error starting service: %v", err)
   }
   ```

5. **Define Actions**:
   Actions represent the core business logic. Example:
   
   ```go
   func ExampleAction(a typ.IAction) any {
     action, _ := a.(*act.Action)
     return fmt.Sprintf("Hello %s!", action.InputString(0))
   }
   ```

---

## **API Documentation** 

API endpoints are automatically generated based on the registered modules and actions. You can invoke any registered action via HTTP requests. For example, the `m_api` module exposes all actions at `/api/<action_name>`.

---

## **Running the Application** 

1. **Start the Service**:
   ```bash
   go run main.go
   ```

2. **Test the API**:
   Access your API at `http://localhost:8080/api/`. Example:

   - `GET /api/example.action` triggers the `example.action` from the `m_example` module.

---

## **Contributing** 

We welcome contributions to the Mini Framework! Here's how you can help:

- **Fork** the repository.
- Create a **new branch** for your changes.
- Submit a **pull request** with a detailed description of the changes.
  
Please include:
- Clear documentation.
- Unit tests to verify the functionality.

---

## **License** 

The Mini Framework is open-source and distributed under the MIT License.

---

For any further questions or clarifications, feel free to reach out. We’re happy to help! 😄