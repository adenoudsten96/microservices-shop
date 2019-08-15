from flask import Flask, render_template, request, jsonify
import time
import uuid
app = Flask(__name__)


@app.route('/email', methods=['POST'])
def index():

	# Simulate an email processing
	time.sleep(1)

  # Make JSON response
	payload = {
		"success": "true",
	}
	payload = jsonify(payload)

	# Return success, email sent
	return payload, 200

if __name__ == '__main__':
  app.run(host='0.0.0.0', port=8002, debug=True)
 