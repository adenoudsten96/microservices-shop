from starlette.applications import Starlette
from starlette.responses import JSONResponse, Response
import uvicorn
import time
import uuid
app = Starlette(debug=False)


@app.route('/email', methods=['POST'])
async def index(request):

	# Simulate an email processing
	time.sleep(1)

    # Make JSON response
	payload = {
		"success": "true",
	}

	# Return success, email sent
	return JSONResponse(payload, 200)

@app.route('/health', methods=['GET'])
async def healthcheck(request):
	return Response("OK", 200)


if __name__ == '__main__':
  uvicorn.run(app, host='0.0.0.0', port=8002)
 