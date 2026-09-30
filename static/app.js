function openDetailModal(btn) {
  document.getElementById('modalNama').textContent = btn.dataset.nama;
  document.getElementById('modalNipNim').textContent = btn.dataset.nipnim;
  document.getElementById('modalStatus').textContent = btn.dataset.status;
  document.getElementById('modalDetail').textContent = btn.dataset.detail || '(tidak ada catatan tambahan)';
  document.getElementById('detailModal').classList.add('open');
}
function closeDetailModal() {
  document.getElementById('detailModal').classList.remove('open');
}

function openEditModal(btn) {
  document.getElementById('editId').value = btn.dataset.id;
  document.getElementById('editStatus').value = btn.dataset.status;
  document.getElementById('editDetail').value = btn.dataset.detail || '';
  document.getElementById('completionContactEmail').value = btn.dataset.contactEmail || '';
  suggestAccountEmail(btn.dataset.nama, btn.dataset.nipnim, btn.dataset.jenis);
  var editForm = document.getElementById('editStatus').form;
  editForm.dataset.requestEmail = getSuggestedAccount(btn.dataset.nama, btn.dataset.nipnim, btn.dataset.jenis);
  if (!document.getElementById('editDetail').value) updateStatusDetailTemplate();
  updateCompletionFields();
  document.getElementById('editModal').classList.add('open');
}

function getSuggestedAccount(name, nipNim, requestType) {
  var namePart = (name || '').toLowerCase().replace(/[^a-z0-9]/g, '');
  if (requestType === 'Mahasiswa ITH') {
    var studentNim = (nipNim || '').trim();
    return namePart && studentNim ? namePart + '.' + studentNim + '@mahasiswa.ith.ac.id' : '';
  }
  return namePart ? namePart + '@ith.ac.id' : '';
}

function suggestAccountEmail(name, nipNim, requestType) {
  var accountInput = document.getElementById('accountEmail');
  var accountLabel = document.getElementById('accountEmailLabel');
  var isStudent = requestType === 'Mahasiswa ITH';
  var namePart = (name || '').toLowerCase().replace(/[^a-z0-9]/g, '');
  var domain = isStudent ? '@mahasiswa.ith.ac.id' : '@ith.ac.id';
  var localPart = isStudent ? namePart + '.' + (nipNim || '').trim() : namePart;

  accountInput.value = localPart && (!isStudent || (nipNim || '').trim()) ? localPart + domain : '';
  accountInput.dataset.suggested = accountInput.value;
  accountLabel.textContent = 'Nama email akun ' + domain;
}
function closeEditModal() {
  document.getElementById('editModal').classList.remove('open');
  document.getElementById('loginPassword').value = '';
}

function updateCompletionFields() {
  var fields = document.getElementById('completionCredentials');
  var required = document.getElementById('editStatus').value === 'Selesai';
  fields.hidden = !required;
  ['completionContactEmail', 'accountEmail', 'loginPassword'].forEach(function (id) {
    document.getElementById(id).required = required;
  });
}

var editStatus = document.getElementById('editStatus');
function updateStatusDetailTemplate() {
  var detail = document.getElementById('editDetail');
  var selectedStatus = document.getElementById('editStatus').value;
  var accountEmail = document.getElementById('editStatus').form.dataset.requestEmail || 'nama@ith.ac.id';

  if (selectedStatus === 'Diproses') {
    detail.value = 'Pengajuan Email dengan nama ' + accountEmail + ' sedang kami proses, mohon untuk memantaunya dalam waktu 2x24 jam';
  } else if (selectedStatus === 'Selesai') {
    detail.value = 'Email yang kamu ajukan sudah jadi dan bisa di cek di email penerima kredensial, Note : jika tidak muncul, masuk pada tab Spam';
  } else if (selectedStatus === 'Ditolak') {
    detail.value = 'Pengajuan akun email kamu kami tolak dengan alasan ...';
  }
}

if (editStatus) {
  editStatus.addEventListener('change', function () {
    updateCompletionFields();
    updateStatusDetailTemplate();
  });
}

document.addEventListener('keydown', function (e) {
  if (e.key === 'Escape') {
    closeDetailModal();
    closeEditModal();
  }
});
document.addEventListener('click', function (e) {
  if (e.target.classList && e.target.classList.contains('modal-overlay')) {
    e.target.classList.remove('open');
  }
});

function updateEmailRecommendation() {
  var jenis = document.getElementById('jenis_usulan');
  var nama = document.getElementById('nama');
  var nipNim = document.getElementById('nip_nim');
  var container = document.getElementById('emailRecommendation');
  var output = document.getElementById('recommendedEmail');
  if (!jenis || !nama || !nipNim || !container || !output) return;

  if (jenis.value !== 'Mahasiswa ITH') {
    container.hidden = true;
    output.textContent = '';
    return;
  }

  var namePart = nama.value.trim().toLowerCase().replace(/[^a-z0-9]/g, '');
  var nimPart = nipNim.value.trim();
  output.textContent = namePart && nimPart ? namePart + '.' + nimPart : 'nama.nim';
  container.hidden = false;
}

['jenis_usulan', 'nama', 'nip_nim'].forEach(function (id) {
  var field = document.getElementById(id);
  if (field) {
    field.addEventListener('input', updateEmailRecommendation);
    if (id === 'jenis_usulan') field.addEventListener('change', updateEmailRecommendation);
  }
});
updateEmailRecommendation();

function refreshRequestResults() {
  var currentResults = document.getElementById('requestResults');
  if (!currentResults || document.hidden || document.querySelector('.modal-overlay.open')) return;

  var active = document.activeElement;
  if (active && /^(INPUT|SELECT|TEXTAREA)$/.test(active.tagName)) return;

  fetch(window.location.href, { cache: 'no-store', credentials: 'same-origin' })
    .then(function (response) {
      if (!response.ok) throw new Error('Gagal memuat pembaruan usulan');
      return response.text();
    })
    .then(function (html) {
      var updatedDocument = new DOMParser().parseFromString(html, 'text/html');
      var updatedResults = updatedDocument.getElementById('requestResults');
      if (!updatedResults || !currentResults.isConnected) return;

      if (currentResults.innerHTML !== updatedResults.innerHTML) {
        currentResults.replaceWith(document.importNode(updatedResults, true));
      }

      var currentTotal = document.getElementById('requestTotal');
      var updatedTotal = updatedDocument.getElementById('requestTotal');
      if (currentTotal && updatedTotal) currentTotal.textContent = updatedTotal.textContent;
    })
    .catch(function () {});
}

window.setInterval(refreshRequestResults, 15000);
