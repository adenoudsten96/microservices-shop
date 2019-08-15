from flask import Flask, render_template, request, jsonify
import time
import random
import uuid
app = Flask(__name__)

class Payment:

	def __init__(self, creditcard, paymentid):
		self.creditcard = creditcard
		self.paymentid = paymentid

payments = []

@app.route('/payment', methods=['POST'])
def index():
	# Get the JSON data
	data = request.get_json()

	# Simulate a payment processing
	time.sleep(1)

	# Save the payment
	id = uuid.uuid4()
	np = Payment(creditcard=data["creditcard"], paymentid=id)
	payments.append(np)

	# Make JSON response
	payload = {
		"transactionid": str(np.paymentid)
	}
	payload = jsonify(payload)

	# Return payment
	return payload, 200


if __name__ == '__main__':
  app.run(host='0.0.0.0', port=8000, debug=True)
 