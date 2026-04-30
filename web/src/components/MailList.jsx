export default function MailList({ mails, selectedMail, onSelect }) {
  if (mails.length === 0) {
    return (
      <div className="card bg-base-100 shadow-md">
        <div className="card-body">
          <h3 className="card-title text-sm">Mails</h3>
          <p className="text-base-content/50 text-sm">No emails yet. Waiting for incoming mail...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="card bg-base-100 shadow-md">
      <div className="card-body p-0">
        <div className="overflow-x-auto">
          <table className="table table-sm">
            <thead>
              <tr>
                <th>From</th>
                <th>Subject</th>
                <th>Time</th>
              </tr>
            </thead>
            <tbody>
              {mails.map((mail, idx) => (
                <tr
                  key={mail.id || idx}
                  className={`cursor-pointer hover:bg-base-200 ${selectedMail === mail ? 'active' : ''}`}
                  onClick={() => onSelect(mail)}
                >
                  <td className="max-w-[120px] truncate">{mail.from}</td>
                  <td className="max-w-[150px] truncate">{mail.subject || '(no subject)'}</td>
                  <td className="whitespace-nowrap">{new Date(mail.created_at).toLocaleTimeString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
