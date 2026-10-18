import { Resend } from "resend";
import { env } from "../config";

const resend = new Resend(env.RESEND_API_KEY);

type Email = {
  to: string;
  subject: string;
  text: string;
  html: string;
};

export async function sendEmail(email: Email): Promise<void> {
  const { error } = await resend.emails.send({ from: env.EMAIL_FROM, ...email });

  if (error) {
    throw new Error(`Failed to send email to ${email.to}: ${error.message}`);
  }
}
