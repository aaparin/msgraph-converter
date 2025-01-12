# DOCX to PDF Converter Service

This service provides DOCX to PDF conversion using Microsoft Graph API and OneDrive. It's built with Go and packaged as a Docker container for easy deployment.

## Features

- Convert DOCX documents to PDF format
- Utilizes Microsoft OneDrive for file processing
- RESTful API endpoints
- Docker support for easy deployment
- Environment-based configuration

## Prerequisites

- Docker installed on your system
- Microsoft Azure account
- Registered application in Azure AD with appropriate permissions

## Setup Azure Application

1. Go to [Azure Portal](https://portal.azure.com)
2. Navigate to Azure Active Directory
3. Register a new application:
   - Select "New Registration"
   - Name your application
   - Select supported account types
   - Click Register
4. Once registered, note down:
   - Application (client) ID
   - Directory (tenant) ID
   - Create a new client secret and save its value

5. Configure API permissions in Azure Portal:
   - Go to your app registration
   - Select "API permissions"
   - Click "Add a permission"
   - Choose "Microsoft Graph"
   - Select "Application permissions"
   - Add the following permissions:
     - Files.ReadWrite.All
     - Sites.ReadWrite.All
     - User.Read.All
   - Click "Grant admin consent" for these permissions
   
Note: These permissions are required for the application to work with files and convert documents on behalf of users.

## Installation

1. Clone the repository:
```bash
git clone https://github.com/aaparin/msgraph-converter.git
cd msgraph-converter
```

2. Copy the environment file:
```bash
cp .env.demo .env
```

3. Update the .env file with your Azure credentials:
```env
AZURE_CLIENT_ID=your_client_id
AZURE_CLIENT_SECRET=your_client_secret
AZURE_TENANT_ID=your_tenant_id
UPLOAD_DIRECTORY=Documents
SERVICE_PORT=8181
```

4. Build the Docker image:
```bash
docker build -t msgraph-converter .
```

5. Run the container:
```bash
docker run -p 8181:8181 --env-file .env msgraph-converter
```

Alternatively, you can run on a different port:
```bash
docker run -p 9000:8181 --env-file .env -e SERVICE_PORT=8181 msgraph-converter
```

## API Endpoints

### List Available Drives

```http
GET http://localhost:8181/drives
```

Response example:
```json
{
    "drives": [
        "b!36******"
    ]
}
```

You can select the desired drive and update it in your .env file.

### Convert DOCX to PDF

```http
POST http://localhost:8181/convert
Content-Type: multipart/form-data
```

Parameters:
- `file`: DOCX file to convert (required)

Response:
- PDF file in the response body

Example using curl:
```bash
curl -X POST -F "file=@document.docx" http://localhost:8181/convert --output result.pdf
```

## Error Handling

The API returns appropriate HTTP status codes and error messages:
- 400: Bad Request (invalid input)
- 401: Unauthorized (invalid credentials)
- 500: Internal Server Error (processing error)

## Security Considerations

- Ensure your .env file is not committed to version control
- The service should be deployed behind a reverse proxy in production
- Use HTTPS in production environments
- Regular updates of dependencies are recommended

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.