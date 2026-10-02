import { NextResponse } from 'next/server';
import crypto from 'crypto';

type IssueRequest = {
  customer_id?: string;
  branch_id?: string;
  industry_id?: string;
  hardware_id?: string;
  order_id?: string;
};

export async function POST(request: Request) {
  try {
    const body = (await request.json()) as IssueRequest;

    const required = [
      'customer_id',
      'branch_id',
      'industry_id',
      'hardware_id',
      'order_id',
    ] as const;

    for (const field of required) {
      if (!body[field]) {
        return NextResponse.json(
          {
            ok: false,
            error: 'missing_required_field',
            field,
            message: `${field} is required`,
          },
          { status: 400 }
        );
      }
    }

    const licenseId =
      'HZ-' +
      crypto.randomBytes(8).toString('hex').toUpperCase();

    return NextResponse.json({
      ok: true,
      license_id: licenseId,
      customer_id: body.customer_id,
      branch_id: body.branch_id,
      industry_id: body.industry_id,
      hardware_id: body.hardware_id,
      order_id: body.order_id,
      mode: 'demo',
    });
  } catch {
    return NextResponse.json(
      {
        ok: false,
        error: 'invalid_json',
        message: 'Invalid JSON request',
      },
      { status: 400 }
    );
  }
}
