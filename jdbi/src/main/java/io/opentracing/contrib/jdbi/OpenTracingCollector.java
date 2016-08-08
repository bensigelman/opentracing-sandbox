package io.opentracing.contrib.jdbi;

import io.opentracing.Span;
import io.opentracing.Tracer;
import org.skife.jdbi.v2.SQLStatement;
import org.skife.jdbi.v2.StatementContext;
import org.skife.jdbi.v2.TimingCollector;

/**
 * XXX: comment
 */
public class OpenTracingCollector implements TimingCollector {
    public final static String PARENT_SPAN_ATTRIBUTE_KEY = "io.opentracing.parent";

    private final Tracer tracer;

    public OpenTracingCollector(Tracer tracer) {
        this.tracer = tracer;
    }

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

    public static void setParent(SQLStatement<?> statement, Span span) {
        statement.getContext().setAttribute(PARENT_SPAN_ATTRIBUTE_KEY, span);
    }
}
