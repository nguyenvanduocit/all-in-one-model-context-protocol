Here's the English translation of the README.md:

# Guide to Getting token.json from Google API

## 0. Create a project on Google Cloud Platform

1. Access Google Cloud Console (https://console.cloud.google.com/)
2. Sign in with your Google account
3. Create a new project:
   - Click on the dropdown menu in the top left corner (next to "Google Cloud")
   - Click "New Project"
   - Enter project name
   - Select organization (if applicable)
   - Click "Create"

4. Enable Gmail API:
   - From the left menu, select "APIs & Services" > "Library"
   - Search for "Gmail API"
   - Click on Gmail API
   - Click "Enable"

5. Configure OAuth consent screen:
   - From "APIs & Services" menu, select "OAuth consent screen"
   - Choose User Type as "External" (or "Internal" if you use Google Workspace)
   - Click "Create"
   - Fill in required information:
     + App name: Your application name
     + User support email: Contact email
     + Developer contact information: Contact email
   - Click "Save and Continue"
   - In the Scopes section, click "Add or Remove Scopes"
   - Add necessary scopes (e.g., Gmail API scopes)
   - Click "Save and Continue"
   - Add test users if needed
   - Click "Save and Continue"

## 1. Preparation

1. Ensure Go is installed on your computer
2. Create a new directory for the project
3. Copy `main.go` file into the project directory

## 2. Get credentials.json

1. Access Google Cloud Console
2. Create a new project or select an existing one
3. In the menu, select "APIs & Services" > "Credentials"
4. Click "Create Credentials" > "OAuth client ID"
5. Choose Application type as "Desktop app"
6. Name your credential and click "Create"
7. Download credentials.json and place it in the same directory as main.go

## 3. Run the program

1. Open terminal and cd to the directory containing main.go
2. Run the command:
```bash
go run main.go -credentials=/path/to/google-credentials.json -token=/path/to/google-token.json
```

Remember the paths, because you will need them in the next step.

## 4. Authenticate and get token

1. The program will display a URL. Copy this URL
2. Open the URL in your browser
3. Sign in to your Google account
4. Accept the requested access permissions
5. Google will provide an authorization code
6. Copy this authorization code
7. Return to terminal and paste the authorization code
8. Press Enter

## 5. Results

- The program will automatically create `token.json` file in the current directory
- This token.json contains access token and refresh token
- You can use this token for subsequent API access
- The program will display the list of labels in your Gmail

## 6. Configure the MCP

Set the path of two file as following key in your claude config file:

```
{
   ...
   "GOOGLE_CREDENTIALS_FILE": "/path/to/google-credentials.json",
   "GOOGLE_TOKEN_FILE": "/path/to/google-token.json",
   ...
}
```

## Notes

- token.json contains sensitive information, don't share it
- Tokens have expiration dates but will auto-refresh
- If you change scopes in the code, you need to delete the old token.json and create a new one

## Common Error Handling

1. If unable to read credentials.json:
   - Check if the file exists in the directory
   - Verify the filename is exactly "credentials.json"

2. If authorization code is invalid:
   - Ensure you copied the code correctly and completely
   - Try generating a new code

3. If access is denied:
   - Check scopes in the code
   - Confirm Gmail API is enabled in Google Cloud Console
