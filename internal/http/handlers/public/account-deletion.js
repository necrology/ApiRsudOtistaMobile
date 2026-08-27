const requestForm = document.querySelector('#request-form');
const confirmForm = document.querySelector('#confirm-form');
const requestStatus = document.querySelector('#request-status');
const confirmStatus = document.querySelector('#confirm-status');
const requestIdentifier = document.querySelector('#request-identifier');
const confirmIdentifier = document.querySelector('#confirm-identifier');

function showStatus(element, message, success) {
  element.textContent = message;
  element.className = `status visible ${success ? 'success' : 'error'}`;
}

async function postJSON(path, body) {
  const response = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
    body: JSON.stringify(body),
  });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok || payload.success === false) {
    throw new Error(payload.message || 'Permintaan tidak dapat diproses.');
  }
  return payload;
}

requestForm.addEventListener('submit', async (event) => {
  event.preventDefault();
  const button = requestForm.querySelector('button');
  button.disabled = true;
  requestStatus.className = 'status';

  try {
    const identifier = requestIdentifier.value.trim();
    const payload = await postJSON('/api/v1/auth/account-deletion/web/request', {
      identifier,
      password: document.querySelector('#request-password').value,
    });
    confirmIdentifier.value = identifier;
    requestForm.reset();
    showStatus(requestStatus, payload.message, true);
    document.querySelector('#confirm-otp').focus();
  } catch (error) {
    showStatus(requestStatus, error.message, false);
  } finally {
    button.disabled = false;
  }
});

confirmForm.addEventListener('submit', async (event) => {
  event.preventDefault();
  const button = confirmForm.querySelector('button');
  button.disabled = true;
  confirmStatus.className = 'status';

  try {
    const payload = await postJSON('/api/v1/auth/account-deletion/web/confirm', {
      identifier: confirmIdentifier.value.trim(),
      otp: document.querySelector('#confirm-otp').value.trim(),
    });
    confirmForm.reset();
    showStatus(confirmStatus, `${payload.data?.message || 'Akun berhasil dihapus.'} ${payload.data?.retained_data || ''}`.trim(), true);
  } catch (error) {
    showStatus(confirmStatus, error.message, false);
  } finally {
    button.disabled = false;
  }
});
