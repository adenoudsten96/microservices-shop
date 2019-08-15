from flask import Flask, render_template, request, jsonify
import time
import random
app = Flask(__name__)

class Payment:

	def __init__(self, creditcard, paymentid):
		self.creditcard = creditcard
		self.paymentid = id

payments = []

@app.route('/payment', methods=['POST'])
def index():
	# Get the JSON data
	data = request.get_json()

	# Simulate a payment processing
	time.sleep(1)

	# Save the payment
	np = Payment(creditcard=data["creditcard"], paymentid=random.randint(1, 100000))
	payments.append(np)

	# Make JSON response
	payload = {
		"status": "success"
	}
	payload = jsonify(payload)

	# Return payment
	return payload, 201


if __name__ == '__main__':
  app.run(host='0.0.0.0', port=8000, debug=True)
 