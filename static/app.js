document.addEventListener('DOMContentLoaded', () => {
     const loggedOutView = document.getElementById('logged-out-view');
     const loggedInView = document.getElementById('logged-in-view');
     const userInfoDiv = document.getElementById('user-info');
     const invoicesTableBody = document.querySelector('#invoices-table tbody');

     const checkAuthStatus = async () => {
         try {
             const response = await fetch('/api/v1/auth/status');
             if (response.ok) {
                 const user = await response.json();
                 showLoggedInView(user);
             } else {
                 showLoggedOutView();
             }
         } catch (error) {
             console.error('Error checking auth status:', error);
             showLoggedOutView();
         }
     };

     const showLoggedOutView = () => {
         loggedInView.style.display = 'none';
         loggedOutView.style.display = 'block';
         userInfoDiv.innerHTML = '';
     };

     const showLoggedInView = (user) => {
         loggedOutView.style.display = 'none';
         loggedInView.style.display = 'block';
         userInfoDiv.innerHTML = `<p>Logged in as: ${user.email}</p><button
      id="logout-button">Logout</button>`;

         document.getElementById('logout-button').addEventListener('click', logout);
         fetchStagedInvoices();
     };

     const fetchStagedInvoices = async () => {
         try {
             const response = await fetch('/api/v1/invoices/staged');
             const invoices = await response.json();
             invoicesTableBody.innerHTML = ''; // Clear existing rows
             if (invoices && invoices.length > 0) {
                 invoices.forEach(invoice => {
                     const row = document.createElement('tr');
                     row.innerHTML = `
                         <td>${new Date(invoice.ReceivedAt * 1000).toLocaleDateString
      ()}</td>
                         <td>${invoice.Sender}</td>
                         <td>${invoice.Subject}</td>
                         <td class="actions">
                             <button class="approve-btn" data-id="${invoice.ID}">Approve
      </button>
                             <button class="reject-btn" data-id="${invoice.ID}">Reject
      </button>
                         </td>
                     `;
                     invoicesTableBody.appendChild(row);
                 });
             } else {
                 invoicesTableBody.innerHTML = '<tr><td colspan="4">No invoices pending review.</td></tr>';
             }
         } catch (error) {
             console.error('Error fetching invoices:', error);
         }
     };

     const logout = async () => {
         await fetch('/api/v1/auth/logout', { method: 'POST' });
         showLoggedOutView();
     };

     // Event delegation for approve/reject buttons
     invoicesTableBody.addEventListener('click', async (event) => {
         const target = event.target;
         const invoiceId = target.dataset.id;
         if (!invoiceId) return;

         if (target.classList.contains('approve-btn')) {
             await fetch(`/api/v1/invoices/${invoiceId}/approve`, { method: 'POST' });
             target.closest('tr').remove(); // Remove from UI on success
         } else if (target.classList.contains('reject-btn')) {
             await fetch(`/api/v1/invoices/${invoiceId}/reject`, { method: 'POST' });
             target.closest('tr').remove();
         }
     });

     // Initial check
     checkAuthStatus();
 });