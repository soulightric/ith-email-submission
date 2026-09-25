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
  document.getElementById('editModal').classList.add('open');
}
function closeEditModal() {
  document.getElementById('editModal').classList.remove('open');
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

  var namePart = nama.value.trim().toLowerCase().replace(/\s+/g, '');
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
