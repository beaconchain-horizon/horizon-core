export type AgentStatus = 'online' | 'offline' | 'warning';

export const angelAgents = Array.from({ length: 144 }, (_, i) => {
  const n = String(i + 1).padStart(3, '0');

  return {
    id: `agent-${n}`,
    name: `Horizon Agent ${n}`,
    site: `Site ${n}`,
    status:
      i % 11 === 0
        ? 'offline'
        : i % 7 === 0
          ? 'warning'
          : 'online',
  } satisfies {
    id: string;
    name: string;
    site: string;
    status: AgentStatus;
  };
});

export const angelCustomers = [
  {
    id: 'customer-demo-01',
    name: 'Demo Customer',
    type: 'Industrial',
    branches: 1,
    status: 'active',
  },
  {
    id: 'customer-demo-02',
    name: 'Demo Bank',
    type: 'Bank',
    branches: 2,
    status: 'active',
  },
];

export const angelDemoOrders = [
  {
    id: 'ORD-DEMO-001',
    customer: 'Demo Customer',
    product: 'Industrial Monitoring',
    qty: 1,
    status: 'pending_issuance',
  },
  {
    id: 'ORD-DEMO-002',
    customer: 'Demo Bank',
    product: 'Bank Settlement',
    qty: 2,
    status: 'pending_issuance',
  },
];
