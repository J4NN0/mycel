# Resend

Mycel can send emails on your behalf through [Resend](https://resend.com). This is optional: the `send_email` tool is only registered when **both** `RESEND_API_KEY` and `RESEND_FROM` are set. Otherwise, the agent starts without it and simply cannot send mail.

## Get an API key

1. Sign up at [resend.com](https://resend.com).
2. Create an API key (Dashboard → API Keys) and set it as `RESEND_API_KEY`.

## Choose a sender address

Set the address the mail is sent from as `RESEND_FROM`:

=== "Quick testing"

    Use the shared sandbox sender — no setup required:

    ```dotenv
    RESEND_FROM=onboarding@resend.dev
    ```

    It can only send to the email address you signed up to Resend with.

=== "Real recipients"

    Verify a domain you own (Dashboard → Domains, then add the SPF/DKIM records) and use an address
    on that domain:

    ```dotenv
    RESEND_FROM=agent@yourdomain.com
    ```

## Using it

Just ask, in the terminal or on Telegram:

> Send an email to sam@example.org saying I'll be late for the standup.

The tool is deliberately conservative: it only fires when your latest message explicitly asks for an email, it refuses placeholder recipients on reserved domains such as `example.com`, and it will ask you for a real address rather than invent one.

!!! note "The model needs tool support"
    Tool calling depends on the model. If `LLM_MODEL` has no `tools` capability, Mycel warns at
    startup and the email tool is ignored no matter how it is configured.
