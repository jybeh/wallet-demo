INSERT INTO account (
    account_id,
    name,
    type,
    currency,
    balance,
    created_at,
    updated_at
) VALUES (
             '1000000001',
             'Holding Account',
             'wallet',
             'MYR',
             100000000000,
             NOW(),
             NOW()
         );

INSERT INTO account (
    account_id,
    name,
    type,
    currency,
    balance,
    created_at,
    updated_at
) VALUES (
             '12345678',
             'Demo Wallet 1',
             'wallet',
             'MYR',
             100000,             -- RM 1,000.00 in minor unit (e.g., sen)
             NOW(),
             NOW()
         );

INSERT INTO account (
    account_id,
    name,
    type,
    currency,
    balance,
    created_at,
    updated_at
) VALUES (
             '87654321',
             'Demo Wallet 2',
             'wallet',
             'MYR',
             100000,             -- RM 1,000.00 in minor unit (e.g., sen)
             NOW(),
             NOW()
         );