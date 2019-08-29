# Check if run as admin
If (-NOT ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator"))

{   
$arguments = "& '" + $myinvocation.mycommand.definition + "'"
Start-Process powershell -Verb runAs -ArgumentList $arguments
Break
}

# Set ENV variabales
[System.Environment]::SetEnvironmentVariable("REDIS_HOST", "localhost:32768", [System.EnvironmentVariableTarget]::User)
[System.Environment]::SetEnvironmentVariable("DB_HOST", "localhost", [System.EnvironmentVariableTarget]::User)
[System.Environment]::SetEnvironmentVariable("DB_PASS", "", [System.EnvironmentVariableTarget]::User)

# Services
[System.Environment]::SetEnvironmentVariable("CHECKOUTSERVICE", "http://localhost:8080", [System.EnvironmentVariableTarget]::User)
[System.Environment]::SetEnvironmentVariable("CARTSERVICE", "http://localhost:8081", [System.EnvironmentVariableTarget]::User)
[System.Environment]::SetEnvironmentVariable("PRODUCTSERVICE", "http://localhost:8082", [System.EnvironmentVariableTarget]::User)
[System.Environment]::SetEnvironmentVariable("PAYMENTSERVICE", "http://localhost:8000", [System.EnvironmentVariableTarget]::User)
[System.Environment]::SetEnvironmentVariable("SHIPPINGSERVICE", "http://localhost:8001", [System.EnvironmentVariableTarget]::User)
[System.Environment]::SetEnvironmentVariable("EMAILSERVICE", "http://localhost:8002", [System.EnvironmentVariableTarget]::User)

# Set the path
$basedir = 'C:\Users\alexd\Desktop\microservices-shop\services'
$directories = Get-ChildItem -Path $basedir | ?{$_.PSIsContainer}
Set-Location $basedir

foreach ($directory in $directories) {

    # cd into dir
    Set-Location -Path $directory

    # get the python or go file
    $pyfile = Get-ChildItem -Filter *.py -Path .\app -ErrorAction silentlycontinue

    if ($pyfile -eq $null) {
        # It's a GO file
        start main.exe
    } else {
        # It's a Python file
        Set-Location .\app
        start python.exe $pyfile
    }

    # change back to base directory
    Set-Location $basedir
}
