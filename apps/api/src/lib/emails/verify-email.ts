type VerifyEmailInput = {
  name: string;
  url: string;
};

export function verifyEmailTemplate({ name, url }: VerifyEmailInput) {
  return {
    subject: "Verify your email for Kyno",
    text: `Hi ${name},\n\nConfirm your email to start your cinema archive:\n${url}\n\nIf you didn't create a Kyno account, you can ignore this email.`,
    html: `
      <p>Hi ${escapeHtml(name)},</p>
      <p>Confirm your email to start your cinema archive.</p>
      <p><a href="${url}">Verify email</a></p>
      <p>If you didn't create a Kyno account, you can ignore this email.</p>
    `,
  };
}

function escapeHtml(value: string): string {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}
