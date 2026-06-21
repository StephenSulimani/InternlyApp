import { type ReactNode } from "react";
import ScrollWatermark from "./ScrollWatermark";
import styles from "./PageShell.module.css";

type Props = {
  children: ReactNode;
};

export default function PageShell({ children }: Props) {
  return (
    <div className={styles.shell}>
      <ScrollWatermark />
      <div className={styles.content}>{children}</div>
    </div>
  );
}
