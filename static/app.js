document.addEventListener('DOMContentLoaded', () => {
     const loggedOutView = document.getElementById('logged-out-view');
     const loggedInView = document.getElementById('logged-in-view');
     const userInfoDiv = document.getElementById('user-info');
     const invoicesTableBody = document.querySelector('#invoices-table tbody');
     const prevPageBtn = document.getElementById('prev-page-btn');
     const nextPageBtn = document.getElementById('next-page-btn');
     const pageInfoSpan = document.getElementById('page-info');
     const notificationArea = document.getElementById('notification-area');

     let currentPage = 1;
     const limit = 25;

     const showNotification = (message, type = 'success') => {
        notificationArea.innerHTML = `
        <div class="notification ${type}">
            <span>${message}</span>
            <button class="close-btn">&times;</button>
        </div>
        `; 

        const closeButton = notificationArea.querySelector('.close-btn');
        const notificationDiv = notificationArea.querySelector('.notification');

        const closeNotification = () => {
            if (notificationArea) {
                notificationArea.innerHTML = '';
            }
        };

        closeButton.addEventListener('click', closeNotification);

        setTimeout(closeNotification, 5000);
     };

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
         userInfoDiv.innerHTML = `<p>Logged in as: ${user.Email}</p><button
      id="logout-button">Logout</button>`;

         document.getElementById('logout-button').addEventListener('click', logout);
         fetchStagedInvoices();
     };

     const fetchStagedInvoices = async () => {
        const offset = (currentPage - 1) * limit;
         try {
             const response = await fetch(`/api/v1/invoices/staged?limit=${limit}&offset=${offset}`);
             const invoices = await response.json();
             invoicesTableBody.innerHTML = ''; 
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
        
             pageInfoSpan.textContent = `Page ${currentPage}`;
             prevPageBtn.disabled = currentPage === 1;
             nextPageBtn.disabled = !invoices;

         } catch (error) {
             console.error('Error fetching invoices:', error);
         }
     };

     prevPageBtn.addEventListener(`click`, () => {
        if (currentPage > 1) {
            currentPage--;
            fetchStagedInvoices();
        }
     });

     nextPageBtn.addEventListener(`click`, () => {
        currentPage++;
        fetchStagedInvoices();
     });

     const logout = async () => {
         await fetch('/api/v1/auth/logout', { method: 'POST' });
         showLoggedOutView();
     };

     invoicesTableBody.addEventListener('click', async (event) => {
         const target = event.target;
         const invoiceId = target.dataset.id;
         if (!invoiceId) return;

         let response;
         let action;

         if (target.classList.contains('approve-btn')) {
             response = await fetch(`/api/v1/invoices/${invoiceId}/approve`, { method: 'POST' });
             action = 'approved';
         } else if (target.classList.contains('reject-btn')) {
             response = await fetch(`/api/v1/invoices/${invoiceId}/reject`, { method: 'POST' });
             action = 'rejected';
         } else {
            return;
         }
         if (response.ok) {
            showNotification(`Invoice successfully ${action}!`, 'success');
            target.closest('tr').remove();
         } else {
            const errorData = await response.json();
            showNotification(`Error: ${errorData.error}`, 'error')
         }
     });

     checkAuthStatus();
 });