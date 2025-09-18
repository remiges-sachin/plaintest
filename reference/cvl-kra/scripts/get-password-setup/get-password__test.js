// Parse password from SOAP response and store in environment
if (pm.response.code === 200) {
    const responseXml = pm.response.text();
    const xml2js = require('xml2js');
    const parseString = xml2js.parseString;

    parseString(responseXml, { explicitArray: false }, (err, result) => {
        if (err) {
            console.error('Error parsing XML:', err);
            return;
        }

        try {
            // Handle different SOAP envelope formats
            let password;
            if (result['soap:Envelope']) {
                password = result['soap:Envelope']['soap:Body']['GetPasswordResponse']['GetPasswordResult'];
            } else if (result['soap12:Envelope']) {
                password = result['soap12:Envelope']['soap12:Body']['GetPasswordResponse']['GetPasswordResult'];
            }

            if (password) {
                pm.environment.set('password', password);
                console.log('[PASSWORD] Password retrieved and stored');

                pm.test('Password service available', function () {
                    pm.expect(password).to.be.a('string');
                    pm.expect(password.length).to.be.greaterThan(0);
                });
            }
        } catch (error) {
            console.error('Error extracting password:', error);
        }
    });
} else {
    console.log('Password service returned error:', pm.response.code);
    console.log('Response:', pm.response.text());

    // Check if password already exists in environment
    const existingPassword = pm.environment.get('password');
    if (existingPassword) {
        console.log('Using existing password from environment');
        pm.test('Using existing password', function () {
            pm.expect(existingPassword).to.be.a('string');
        });
    } else {
        pm.test('Password retrieval failed', function () {
            pm.expect.fail('Unable to retrieve password and no existing password found');
        });
    }
}
