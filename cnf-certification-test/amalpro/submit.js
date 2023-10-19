/** @param {HTMLFormElement} form */
export async function submit(form) {
    form.elements.submit.disabled = true;
    const formdata = new FormData(form);
    for (const el of form.elements) if (el instanceof HTMLFieldSetElement) el.disabled = true
  
    // Collect data from form fields
    const fields = Array.from(formdata.entries()).reduce((acc, [key, val]) => ({ ...acc,
      [key]: key in acc ?  [acc[key], val] : val
    }), {});
  
    delete fields.submit;
    console.log(fields);
  
    // Send an HTTP request to the server to run the function
    let heading;
    let message;
    let state = 'success';
  
    try {
      const data = await fetch('/runFunction', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', },
        body: JSON.stringify(fields),
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
      message = error.message;
      state = 'danger';
    } finally {
      form.elements.submit.disabled = false;
      for (const el of form.elements) if (el instanceof HTMLFieldSetElement) el.disabled = false
    }
  
    return { heading, message, state };
  }