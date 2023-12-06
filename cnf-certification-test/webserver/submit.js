/** @param {HTMLFormElement} form */
export async function submit(form) {
  form.elements.submit.disabled = true;
  const formdata = new FormData(form);
  // Iterate over form elements and add those with non-empty values to FormData
  Array.from(form.elements).forEach(element => { // this loop for the checkbox values
    if (element.hasAttribute('value') && element.type != "checkbox") {
      formdata.append((element.id.match(/[a-zA-Z]/g) || []).join(''), element.value);
    }
  });
  for (const el of form.elements) if (el instanceof HTMLFieldSetElement) el.disabled = true

  // Collect data from form fields -- will take all the fileds
  const fields = Array.from(formdata.entries()).reduce((acc, [key, val]) => {
    if (acc[key] === undefined) {
      // If the key is not in the accumulator, set it to the value or an array with the value
      acc[key] = [val];
    } else if (Array.isArray(acc[key])) {
      // If the key is already an array, push the new value to the array
      acc[key].push(val);
    } else {
      // If the key is a single value, convert it to an array with both values
      acc[key] = [acc[key], val];
    }
    return acc;
  }, {});

  delete fields.submit; // delete unnrrd data
  console.log(fields);
  formdata.append("jsonData", JSON.stringify(fields));

  // Send an HTTP request to the server to run the function
  let heading;
  let message;
  let state = 'success';

  try { // post request with the collected data
    const data = await fetch('/runFunction', {
      method: 'POST',
      body: formdata,
    }).then(response => {
      if (response.ok) {
        return response.json();
      } else {
        throw new Error(response.statusText);
      }
    });

    heading = 'Success';
    message = data.Message;

    console.log(data);
  } catch (error) {
    console.error(error);
    heading = 'Error'
    message = error.message+" Run Certification test failed, please check the logs";
    state = 'danger';
  } finally {
    form.elements.submit.disabled = false;
    for (const el of form.elements) if (el instanceof HTMLFieldSetElement) el.disabled = false
  }

  return { heading, message, state };
}