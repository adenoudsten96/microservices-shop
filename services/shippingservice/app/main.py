from starlette.applications import Starlette
from starlette.responses import JSONResponse, Response
import uvicorn
import time
import uuid
app = Starlette(debug=False)


@app.route('/ship', methods=['POST'])
async def index(request):
    # Simulate a shipment processing
    time.sleep(1)

    # Make JSON response
    id = uuid.uuid4()
    payload = {
        "shippingid": str(id)
    }

    # Return success, payment processed
    return JSONResponse(payload, 200)

@app.route('/health', methods=['GET'])
async def healthcheck(request):
	return Response("OK", 200)


if __name__ == '__main__':
  uvicorn.run(app, host='0.0.0.0', port=8001)
