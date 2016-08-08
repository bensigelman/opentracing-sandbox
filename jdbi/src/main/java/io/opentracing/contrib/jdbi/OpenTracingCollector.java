package io.opentracing.contrib.jdbi;

import io.opentracing.Span;
import io.opentracing.Tracer;
import org.skife.jdbi.v2.SQLStatement;
import org.skife.jdbi.v2.StatementContext;
import org.skife.jdbi.v2.TimingCollector;

/**
 * OpenTracingCollector is a JDBI TimingCollector that creates OpenTracing Spans for each JDBI SQLStatement.
 *
 * <p>Example usage:
 * <pre>{@code
 * io.opentracing.Tracer tracer = ...;
 * DBI dbi = ...;
 *
 * // One time only: bind OpenTracing to the DBI instance as a TimingCollector.
 * dbi.setTimingCollector(new OpenTracingCollector(tracer));
 *
 * // Elsewhere, anywhere a `Handle` is available:
 * Handle handle = ...;
 * Span parentSpan = ...;  // optional
 *
 * // Create statements as usual with your `handle` instance.
 *  Query<Map<String, Object>> statement = handle.createQuery("SELECT COUNT(*) FROM accounts");
 *
 * // If a parent Span is available, establish the relationship via setParent.
 * OpenTracingCollector.setParent(statement, parent);
 *
 * // Use JDBI as per usual, and Spans will be created for every SQLStatement automatically.
 * List<Map<String, Object>> results = statement.list();
 * }</pre>
 */
public class OpenTracingCollector implements TimingCollector {
    public final static String PARENT_SPAN_ATTRIBUTE_KEY = "io.opentracing.parent";

    private final Tracer tracer;

    public OpenTracingCollector(Tracer tracer) {
        this.tracer = tracer;
    }

    @Override
    public void collect(long elapsedNanos, StatementContext statementContext) {
        long nowMicros = System.currentTimeMillis() * 1000;
        Tracer.SpanBuilder builder = tracer
                .buildSpan("SQL")
                .withStartTimestamp(nowMicros-(elapsedNanos/1000));
        Span parent = (Span)statementContext.getAttribute(PARENT_SPAN_ATTRIBUTE_KEY);
        if (parent != null) {
            builder = builder.asChildOf(parent);
        }
        try (Span collectSpan = builder.start()) {
            collectSpan.log("Raw SQL", statementContext.getRawSql());
        }
    }

    /**
     * Establish an explicit parent relationship for the (child) Span associated with a SQLStatement.
     *
     * @param statement the JDBI SQLStatement which will act as the child of `parent`
     * @param parent the parent Span for `statement`
     */
    public static void setParent(SQLStatement<?> statement, Span parent) {
        statement.getContext().setAttribute(PARENT_SPAN_ATTRIBUTE_KEY, parent);
    }
}
