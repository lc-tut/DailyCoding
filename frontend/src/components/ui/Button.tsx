import styles from "./Button.module.css";

type Variant = "primary" | "secondary";

type Props = React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: Variant;
  loading?: boolean;
};

export default function Button({
  variant = "primary",
  loading = false,
  children,
  disabled,
  className,
  ...rest
}: Props) {
  return (
    <button
      {...rest}
      disabled={disabled || loading}
      className={[styles.btn, styles[variant], className].filter(Boolean).join(" ")}
    >
      {loading ? "実行中..." : children}
    </button>
  );
}
