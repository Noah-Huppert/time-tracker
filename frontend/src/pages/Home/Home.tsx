import { Box, Card, CardContent, CardHeader, Typography } from "@mui/material";
import { Header } from "../../components/Header/Header";
import { Link } from "react-router-dom";
import { ROUTES } from "../../lib/routes";

export const PageHome = () => {
  return (
    <>
      <Header />
      <Box>
        <RoutePaper
          name="Time Entries"
          description="List time entries"
          to={ROUTES.time_entries.make()}
        />

        <RoutePaper
          name="Invoices"
          description="List invoices"
          to={ROUTES.invoices.make()}
        />
      </Box>
    </>
  );
}

const RoutePaper = ({
  name,
  description,
  to,
}: {
  readonly name: string
  readonly description: string
  readonly to: string
}) => {
  return (
    <Link to={to}>
      <Card>
        <CardContent>
          <Typography
            variant="h6"
          >
            {name}
          </Typography>

          <Typography>
            {description}
          </Typography>
        </CardContent>
      </Card>
    </Link>
  );
}