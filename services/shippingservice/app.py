from flask import Flask, render_template, request, jsonify
import time
import uuid
app = Flask(__name__)

class Ship:

  def __init__(self, address, shippingid):
    self.address = address
    self.shippingid = shippingid

ships = []

@app.route('/ship', methods=['POST'])
def index():
	# Get the JSON data
	data = request.get_json()

	# Simulate a shipment processing
	time.sleep(1)

	# Save the shipment
	id = uuid.uuid4()
	ns = Ship(address=data["address"], shippingid=id)
	ships.append(ns)

	# Make JSON response
	payload = {
		"shippingid": str(ns.shippingid)
	}
	payload = jsonify(payload)

	# Return shipment
	return payload, 200

if __name__ == '__main__':
  app.run(host='0.0.0.0', port=8001, debug=True)
 