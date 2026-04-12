INSERT INTO problems (id, title, description, is_active, published_at)
VALUES (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    '二数の和',
    $md$## 問題：二数の和

標準入力から空白区切りで 2 つの整数を受け取り、その和を出力してください。

### 入力形式

```
A B
```

### 出力形式

```
A + B の値
```

### サンプル

| 入力 | 出力 |
|------|------|
| `2 3` | `5` |
| `10 20` | `30` |
$md$,
    true,
    CURRENT_DATE
) ON CONFLICT (id) DO NOTHING;

INSERT INTO test_cases (problem_id, stdin, expected_stdout, order_index, is_sample, score_weight)
VALUES
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '2 3',   '5',  1, true,  1),
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '10 20', '30', 2, false, 1)
ON CONFLICT DO NOTHING;
