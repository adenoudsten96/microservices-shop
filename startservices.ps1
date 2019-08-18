# Set ENV variabales
[Environment]::SetEnvironmentVariable("REDIS_HOST", "localhost:32768")
[Environment]::SetEnvironmentVariable("DB_HOST", "localhost")
[Environment]::SetEnvironmentVariable("DB_PASS", "Appelflap1")

# Services
[Environment]::SetEnvironmentVariable("CHECKOUTSERVICE", "http://localhost:8080")
[Environment]::SetEnvironmentVariable("CARTSERVICE", "http://localhost:8081")
[Environment]::SetEnvironmentVariable("PRODUCTSERVICE", "http://localhost:8082")
[Environment]::SetEnvironmentVariable("PAYMENTSERVICE", "http://localhost:8000")
[Environment]::SetEnvironmentVariable("SHIPPINGSERVICE", "http://localhost:8001")
[Environment]::SetEnvironmentVariable("EMAILSERVICE", "http://localhost:8002")

# Set the path
$basedir = 'C:\Users\alexd\Desktop\microservices-shop\services'
$directories = Get-ChildItem -Path $basedir | ?{$_.PSIsContainer}
Set-Location $basedir

foreach ($directory in $directories) {

    Write-Host $directory
    # cd into dir
    Set-Location -Path $directory

    # get the python or go file
    $pyfile = Get-ChildItem -Filter *.py

    if ($pyfile -eq $null) {
        # It's a GO file
        start main.exe
    }

    # start the python app as a new process
    start python.exe $pyfile
    
    # change back to base directory
    Set-Location $basedir
}